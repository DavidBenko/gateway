package advanced

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/core/request"
	corevm "gateway/core/vm"
	"gateway/model"
	"io"
	"sync/atomic"

	"github.com/robertkrimen/otto"
)

// RequestPreparer is a function that will turn the json.RawMessage and the RemoteEndpoint model into an initialized
// request.
type RequestPreparer func(*model.RemoteEndpoint, *json.RawMessage, map[int64]io.Closer) (request.Request, error)

func IncludePerform(vm *otto.Otto, accountID int64, endpointSource corevm.DataSource,
	prepare RequestPreparer, pauseTimeout *uint64) {
	vm.Set("_perform", func(call otto.FunctionCall) otto.Value {
		if pauseTimeout != nil {
			atomic.StoreUint64(pauseTimeout, 1)
			defer atomic.StoreUint64(pauseTimeout, 0)
		}

		connections := make(map[int64]io.Closer)

		r, err := corevm.GetArgument(call, 0)
		if err != nil {
			return corevm.OttoErrorObject(vm, "invalid argument")
		}

		results, _, err := perform(r, vm, endpointSource, accountID, prepare, connections)
		if err != nil {
			return corevm.OttoErrorObject(vm, err.Error())
		}

		sr, err := json.Marshal(results)
		if err != nil {
			return corevm.OttoErrorObject(vm, err.Error())
		}

		return corevm.ToOttoObjectValue(vm, string(sr))
	})

	scripts := []string{
		"var AP = AP || {}",
		"AP.Perform = function() {return _perform.apply(this, arguments)};",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}

func parseOptions(o map[string]interface{}) (*requestOptions, error) {
	jsonOptions, err := json.Marshal(o)
	if err != nil {
		return nil, errors.New("unable to parse options object")
	}

	var options *requestOptions
	err = json.Unmarshal(jsonOptions, &options)
	if err != nil {
		return nil, errors.New("invalid request options object")
	}

	// Create json.RawMessage from the requests
	options.rawRequests = make([]*json.RawMessage, len(options.Requests))
	for i, request := range options.Requests {
		r, err := json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("invalid request at index %d", i)
		}

		var raw *json.RawMessage
		err = json.Unmarshal(r, &raw)
		if err != nil {
			return nil, err
		}

		options.rawRequests[i] = raw
	}

	return options, nil
}

func perform(o interface{}, vm *otto.Otto, endpointSource corevm.DataSource, accountID int64,
	prepare RequestPreparer, connections map[int64]io.Closer) (interface{}, string, error) {

	switch o.(type) {
	case map[string]interface{}:
		// single endpoint
		options, err := parseOptions(o.(map[string]interface{}))
		if err != nil {
			return nil, "", err
		}

		results := makeRequests(options, accountID, endpointSource, prepare, connections)

		responses := make(map[string][]request.Response, 1)
		responses[options.Codename] = results

		return responses, options.Codename, nil
	case []map[string]interface{}:
		// multiple endpoints
		allopts := o.([]map[string]interface{})
		c := make(chan *responsesWrapper, len(allopts))

		for i, opt := range allopts {
			go func(index int, options map[string]interface{}) {
				responses, codename, err := perform(options, vm, endpointSource, accountID, prepare, connections)
				if err != nil {
					res := &errorResponse{err.Error()}
					c <- &responsesWrapper{"", []request.Response{res}}
				} else {
					c <- &responsesWrapper{codename, responses.(map[string][]request.Response)[codename]}
				}
			}(i, opt)
		}

		responses := make(map[string][]request.Response, len(allopts))
		for i := 0; i < len(allopts); i++ {
			select {
			case result := <-c:
				if result != nil {
					if val, ok := responses[result.codename]; ok {
						responses[result.codename] = append(val, result.responses...)
					} else {
						responses[result.codename] = result.responses
					}
				}
			}
		}

		return responses, "", nil
	default:
		return nil, "", errors.New("invalid argument type")
	}
}

func makeRequests(options *requestOptions, accountID int64, endpointSource corevm.DataSource,
	prepare RequestPreparer, connections map[int64]io.Closer) []request.Response {

	criteria := &corevm.RemoteEndpointStoreCriteria{AccountID: accountID, Codename: options.Codename}
	result, ok := endpointSource.Get(criteria)
	if !ok {
		return []request.Response{&errorResponse{fmt.Sprintf("could not find remote endpoint %s", options.Codename)}}
	}

	endpoint := result.(*model.RemoteEndpoint)

	responses := make([]request.Response, len(options.rawRequests))

	c := make(chan *responseWrapper, len(options.rawRequests))
	for i, r := range options.rawRequests {
		go func(index int, req *json.RawMessage) {
			preparedRequest, err := prepare(endpoint, req, connections)
			if err != nil {
				e := &errorResponse{err.Error()}
				c <- &responseWrapper{index, e}
			} else {
				c <- &responseWrapper{index, preparedRequest.Perform()}
			}
		}(i, r)
	}

	for i := 0; i < len(options.rawRequests); i++ {
		select {
		case response := <-c:
			responses[response.index] = response.response
		}
	}

	return responses
}

type requestOptions struct {
	Codename    string                   `json:"codename"`
	Requests    []map[string]interface{} `json:"requests"`
	rawRequests []*json.RawMessage
}

type responseWrapper struct {
	index    int
	response request.Response
}

type responsesWrapper struct {
	codename  string
	responses []request.Response
}

// errorResponse is used to wrap an error that may have occurred before the actual
// request was made and before a response was returned.
type errorResponse struct {
	Error string `json:"error"`
}

// JSON satisfies the request.Response interface
func (e *errorResponse) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// Log satisfies the request.Response interface
func (e *errorResponse) Log() string {
	return ""
}
