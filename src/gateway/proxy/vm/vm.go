package vm

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gateway/config"
	"gateway/proxy/requests"

	"github.com/gorilla/sessions"

	"github.com/robertkrimen/otto"

	// Add underscore.js functionality to our VMs
	_ "github.com/robertkrimen/otto/underscore"
)

// ProxyVM is an Otto VM with some helper data stored alongside it.
type ProxyVM struct {
	*otto.Otto
	conf                    config.ProxyServer
	requestID               string
	ProxiedRequestsDuration time.Duration

	/* TODO: Do both of the following get removed? */
	scripts           map[string]*otto.Script
	includedLibraries []string

	w            http.ResponseWriter
	r            *http.Request
	sessionStore *sessions.CookieStore
}

// NewVM returns a new Otto VM initialized with Gateway JavaScript libraries.
func NewVM(
	requestID string,
	w http.ResponseWriter,
	r *http.Request,
	conf config.ProxyServer,
	proxyScripts map[string]*otto.Script,
) (*ProxyVM, error) {

	var files = []string{
		"gateway.js",
		"environments.js",
		"sessions.js",
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

	vm := &ProxyVM{
		otto.New(),
		conf, requestID, 0,
		proxyScripts, []string{},
		w, r,
		nil,
	}

	/* FIXME: Need to move keys to Environment for multi-tenant, not config */
	if conf.AuthKey != "" {
		sessionConfig := [][]byte{[]byte(conf.AuthKey)}
		if conf.EncryptionKey != "" {
			sessionConfig = append(sessionConfig, []byte(conf.EncryptionKey))
		}
		vm.sessionStore = sessions.NewCookieStore(sessionConfig...)
	}

	/* TODO: Bind to objects? & evaluate usage */
	vm.Set("__ap_log", vm.log)                        /* log("foo") instead? */
	vm.Set("__ap_environment_get", vm.environmentGet) /* env("key") instead? */
	vm.Set("__ap_session_get", vm.sessionGet)
	vm.Set("__ap_session_set", vm.sessionSet)
	vm.Set("__ap_session_is_set", vm.sessionIsSet)
	vm.Set("__ap_session_delete", vm.sessionDelete)
	vm.Set("__ap_session_set_options", vm.sessionSetOptions)
	// vm.Set("include", vm.includeLibrary)         /* include all by default? */
	vm.Set("__ap_makeRequests", vm.makeRequests) /* TODO: remove prols */

	if _, err := vm.RunAll(scripts); err != nil {
		return nil, err
	}

	return vm, nil
}

// RunAll runs all the given scripts
func (p *ProxyVM) RunAll(scripts []interface{}) (value otto.Value, err error) {
	for _, script := range scripts {
		value, err = p.Run(script)
		if err != nil {
			return
		}
	}
	return
}

func (p *ProxyVM) log(call otto.FunctionCall) otto.Value {
	log.Printf("%s [req %s] [user] %v", config.Proxy, p.requestID, call.Argument(0).String())
	return otto.Value{}
}

/** TODO: Probably 86 all this */
// func (p *ProxyVM) includeLibrary(call otto.FunctionCall) otto.Value {
// 	libraryName := call.Argument(0).String()
//
// 	alreadyIncluded := false
// 	for _, name := range p.includedLibraries {
// 		if name == libraryName {
// 			alreadyIncluded = true
// 			break
// 		}
// 	}
// 	if alreadyIncluded {
// 		return otto.Value{}
// 	}
// 	p.includedLibraries = append(p.includedLibraries, libraryName)
//
// 	libraryKey, err := keys.ScriptKeyForCodeString(libraryName)
// 	if err != nil {
// 		runtimeError(fmt.Sprintf("Could not make '%s' to script key", libraryName))
// 	}
//
// 	library, ok := p.scripts[libraryKey]
// 	if !ok {
// 		runtimeError(fmt.Sprintf("There is no library named '%s'", libraryName))
// 	}
//
// 	_, err = p.Run(library)
// 	if err != nil {
// 		runtimeError(fmt.Sprintf("Error in library '%s': %s", libraryName, err))
// 	}
//
// 	return otto.Value{}
// }

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
