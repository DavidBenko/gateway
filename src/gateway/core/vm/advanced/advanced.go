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
	connections := make(map[int64]io.Closer)
	vm.Set("_perform", func(call otto.FunctionCall) otto.Value {

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

		rs, err := json.Marshal(r)
		if err != nil {
			return corevm.OttoErrorObject(vm, err.Error())
		}

		var raw *json.RawMessage
		err = json.Unmarshal(rs, &raw)
		if err != nil {
			return corevm.OttoErrorObject(vm, err.Error())
		}

		criteria := &corevm.RemoteEndpointStoreCriteria{AccountID: accountID, Codename: codename}
		result, ok := endpointSource.Get(criteria)
		if !ok {
			return corevm.OttoErrorObject(vm, fmt.Sprintf("could not find remote endpoint '%s'", codename))
		}

		endpoint := result.(*model.RemoteEndpoint)

		req, err := prepare(endpoint, raw, connections)
		if err != nil {
			return corevm.OttoErrorObject(vm, err.Error())
		}

		response := req.Perform()

		rawResponse, err := response.JSON()
		if err != nil {
			return corevm.OttoErrorObject(vm, err.Error())
		}

		return corevm.ToOttoObjectValue(vm, string(rawResponse))
	})

	scripts := []string{
		"var AP = AP || {}",
		"AP.Perform = function() {return _perform.apply(this, arguments)};",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}
