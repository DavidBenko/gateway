package advanced

import "github.com/robertkrimen/otto"

func IncludePerform(vm *otto.Otto, accountID int64, remoteEndpointSource interface{}) {
	vm.Set("_perform", func(call otto.FunctionCall) otto.Value {
		return otto.Value{}
	})

	scripts := []string{
		"AP.Perform = function() {return _perform(arguments)};",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}
