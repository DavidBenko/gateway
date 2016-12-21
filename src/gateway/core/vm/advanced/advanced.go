package advanced

import (
	"encoding/json"
	"fmt"
	"gateway/core/request"
	corevm "gateway/core/vm"
	"gateway/model"
	"io"

	"github.com/robertkrimen/otto"
)

// RequestPreparer is a function that will turn the json.RawMessage and the RemoteEndpoint model into an initialized
// request.
type RequestPreparer func(*model.RemoteEndpoint, *json.RawMessage, map[int64]io.Closer) (request.Request, error)

func IncludePerform(vm *otto.Otto, accountID int64, endpointSource corevm.DataSource, prepare RequestPreparer) {
	vm.Set("_perform", func(call otto.FunctionCall) otto.Value {
		connections := make(map[int64]io.Closer)

		c, err := corevm.GetArgument(call, 0)
		if err != nil {
			return corevm.OttoErrorObject(vm, "missing codename argument")
		}

		if _, ok := c.(string); !ok {
			return corevm.OttoErrorObject(vm, "codename should be a string")
		}
		codename := c.(string)

		r, err := call.Argument(1).Export()
		if r == nil {
			return corevm.OttoErrorObject(vm, "missing request argument")
		}
		if err != nil {
			return corevm.OttoErrorObject(vm, err.Error())
		}

		switch r.(type) {
		case []map[string]interface{}:
			responses, err := performMultiple(r.([]map[string]interface{}), vm, endpointSource, accountID, codename, prepare, connections)
			if err != nil {
				return corevm.OttoErrorObject(vm, err.Error())
			}

			if len(responses) == 0 {
				return otto.UndefinedValue()
			}

			json := ""
			for _, response := range responses {
				rj, err := response.JSON()
				if err != nil {
					corevm.OttoErrorObject(vm, fmt.Sprintf("could not convert response to Javascript object: %s", err.Error()))
				}
				json += fmt.Sprintf("%s,", rj)
			}

			json = fmt.Sprintf("[%s]", json)
			return corevm.ToOttoObjectValue(vm, json)
		case map[string]interface{}:
			response, err := perform(r.(map[string]interface{}), vm, endpointSource, accountID, codename, prepare, connections)
			if err != nil {
				return corevm.OttoErrorObject(vm, err.Error())
			}

			b, err := response.JSON()
			if err != nil {
				return corevm.OttoErrorObject(vm, err.Error())
			}
			return corevm.ToOttoObjectValue(vm, string(b))
		default:
			return corevm.OttoErrorObject(vm, "invalid request object")
		}
	})

	scripts := []string{
		"var AP = AP || {}",
		"AP.Perform = function() {return _perform.apply(this, arguments)};",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}

func perform(r interface{}, vm *otto.Otto, endpointSource corevm.DataSource, accountID int64,
	codename string, prepare RequestPreparer, connections map[int64]io.Closer) (request.Response, error) {

	rs, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	var raw *json.RawMessage
	err = json.Unmarshal(rs, &raw)
	if err != nil {
		return nil, err
	}

	criteria := &corevm.RemoteEndpointStoreCriteria{AccountID: accountID, Codename: codename}
	result, ok := endpointSource.Get(criteria)
	if !ok {
		return nil, fmt.Errorf("could not find remote endpoint %s", codename)
	}

	endpoint := result.(*model.RemoteEndpoint)

	req, err := prepare(endpoint, raw, connections)
	if err != nil {
		return nil, err
	}

	return req.Perform(), nil
}

type payload struct {
	index    int
	response request.Response
}

// errorResponse is used to wrap an error that may have ocurred before the actual
// request was made and before a response was returned.
type errorResponse struct {
	Error string `json:"error"`
}

func (e *errorResponse) JSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *errorResponse) Log() string {
	return ""
}

func performMultiple(r []map[string]interface{}, vm *otto.Otto, endpointSource corevm.DataSource, accountID int64,
	codename string, prepare RequestPreparer, connections map[int64]io.Closer) ([]request.Response, error) {

	responses := make([]request.Response, len(r))
	c := make(chan *payload, len(r))

	for i, req := range r {
		go func(index int, req interface{}) {
			res, err := perform(req, vm, endpointSource, accountID, codename, prepare, connections)
			if err != nil {
				// return an error response for the consumer to handle
				errResponse := &errorResponse{err.Error()}
				c <- &payload{index, errResponse}
			}
			c <- &payload{index, res}
		}(i, req)
	}

	for i := 0; i < len(r); i++ {
		select {
		case req := <-c:
			responses[req.index] = req.response
		}
	}

	return responses, nil
}
