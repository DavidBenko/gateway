package advanced

import (
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

func IncludePerform(vm *otto.Otto) {
	setPerform(vm)

	scripts := []string{
		"AP.Perform = _perform; delete _perform;",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}

func setPerform(vm *otto.Otto) {
	vm.Set("_perform", func(call otto.FunctionCall) otto.Value {
		logreport.Println("TODO: not implemented")
		return otto.Value{}
	})
}
