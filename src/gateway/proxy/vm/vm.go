package vm

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gateway/config"
	"gateway/db"
	"gateway/model"
	"gateway/proxy/requests"

	"github.com/robertkrimen/otto"

	// Add underscore.js functionality to our VMs
	_ "github.com/robertkrimen/otto/underscore"
)

// ProxyVM is an Otto VM with some helper data stored alongside it.
type ProxyVM struct {
	*otto.Otto
	requestID               string
	ProxiedRequestsDuration time.Duration
	db                      db.DB
}

// NewVM returns a new Otto VM initialized with Gateway JavaScript libraries.
func NewVM(requestID string, db db.DB) (*ProxyVM, error) {
	var files = []string{
		"gateway.js",
		"http/request.js",
		"http/response.js",
	}
	var scripts = make([]interface{}, 0)
	for _, filename := range files {
		fileJS, err := Asset(filename)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, fileJS)
	}

	vm := &ProxyVM{otto.New(), requestID, 0, db}
	vm.Set("__ap_log", vm.log)
	vm.Set("include", vm.include)
	vm.Set("__ap_makeRequests", vm.makeRequests)

	for _, script := range scripts {
		_, err := vm.Run(script)
		if err != nil {
			return nil, err
		}
	}

	return vm, nil
}

func (p *ProxyVM) log(call otto.FunctionCall) otto.Value {
	log.Printf("%s [req %s] [user] %v", config.Proxy, p.requestID, call.Argument(0).String())
	return otto.Value{}
}

func (p *ProxyVM) include(call otto.FunctionCall) otto.Value {
	libraryName := call.Argument(0).String()

	libraryModel, err := p.db.Find(&model.Library{}, "Name", libraryName)
	if err != nil {
		runtimeError(fmt.Sprintf("There is no library named '%s'", libraryName))
	}

	library := libraryModel.(*model.Library)
	_, err = p.Run(library.Script)
	if err != nil {
		runtimeError(fmt.Sprintf("Error in library '%s': %s", libraryName, err))
	}

	return otto.Value{}
}

func (p *ProxyVM) makeRequests(call otto.FunctionCall) otto.Value {
	// Parse requests
	var requestJSONs [][]string
	jsonArg := call.Argument(0).String()
	err := json.Unmarshal([]byte(jsonArg), &requestJSONs)
	if err != nil {
		response := requests.NewErrorResponse(fmt.Errorf(
			"Error '%v' unmarshaling JSON: '%s'\n", err, jsonArg))
		return jsonArrayResponsesValue([]requests.Response{response})
	}
	n := len(requestJSONs)
	requestList := make([]requests.Request, n)
	for i, requestData := range requestJSONs {
		request, err := requests.RequestFromData(requestData)
		if err != nil {
			response := requests.NewErrorResponse(fmt.Errorf(
				"Error '%v' building request from data: '%v'\n",
				err, requestData))
			return jsonArrayResponsesValue([]requests.Response{response})
		}
		requestList[i] = request
	}

	// Make requests
	start := time.Now()
	responses, err := requests.MakeRequests(requestList, p.requestID)
	if err != nil {
		response := requests.NewErrorResponse(fmt.Errorf(
			"Error making requests: %v\n", err))
		return jsonArrayResponsesValue([]requests.Response{response})
	}
	p.ProxiedRequestsDuration += time.Since(start)
	return jsonArrayResponsesValue(responses)
}

func jsonArrayValue(items []string) otto.Value {
	aggregatedJSON := fmt.Sprintf("[%s]", strings.Join(items, ",\n"))
	value, err := otto.ToValue(aggregatedJSON)
	if err != nil {
		log.Fatalf("Could not convert string to JS value: %s", err)
	}
	return value
}

func jsonResponses(responses []requests.Response) []string {
	var jsonResponses = make([]string, len(responses))
	for i, response := range responses {
		jsonResponse, err := response.JSON()
		if err != nil {
			jsonResponse, err = requests.NewErrorResponse(err).JSON()
			if err != nil {
				log.Fatalf("Error getting JSON for error response: %s", err)
			}
		}
		jsonResponses[i] = string(jsonResponse)
	}
	return jsonResponses
}

func jsonArrayResponsesValue(responses []requests.Response) otto.Value {
	return jsonArrayValue(jsonResponses(responses))
}

// Panics with otto.Value are caught as runtime errors.
func runtimeError(err string) {
	errValue, _ := otto.ToValue(err)
	panic(errValue)
}
