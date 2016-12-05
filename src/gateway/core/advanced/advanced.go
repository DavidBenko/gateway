package advanced

import (
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

func IncludePerform() {
	scripts := []string{
		"AP.Perform = _perform; delete _perform;",
	}
}

func setPerform(vm *otto.Otto) {
	vm.Set("_perform", func(call otto.FunctionCall) otto.Value {
		logreport.Println("TODO: not implemented")
	})
}
