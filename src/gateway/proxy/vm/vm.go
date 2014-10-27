package vm

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"gateway/proxy/requests"

	"github.com/robertkrimen/otto"
)

// NewVM returns a new Otto VM initialized with Gateway JavaScript libraries.
func NewVM() (*otto.Otto, error) {
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

	vm := otto.New()
	vm.Set("__ap_log", vmLog)
	vm.Set("__ap_makeRequests", makeRequests)

	for _, script := range scripts {
		_, err := vm.Run(script)
		if err != nil {
			return nil, err
		}
	}

	return vm, nil
}

func vmLog(call otto.FunctionCall) otto.Value {
	log.Println(call.Argument(0).String())
	return otto.Value{}
}

func makeRequests(call otto.FunctionCall) otto.Value {
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
	responses, err := requests.MakeRequests(requestList)
	if err != nil {
		response := requests.NewErrorResponse(fmt.Errorf(
			"Error making requests: %v\n", err))
		return jsonArrayResponsesValue([]requests.Response{response})
	}

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
