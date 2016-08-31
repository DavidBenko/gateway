package conversion

import (
	"github.com/robertkrimen/otto"
)

var undefined = otto.Value{}

func IncludeConversion(vm *otto.Otto) {
	setToJSON(vm)
	setToXML(vm)
}

func setToJSON(vm *otto.Otto) {
	vm.Set("_toJson", func(call otto.FunctionCall) otto.Value {
		return undefined
	})
}

func setToXML(vm *otto.Otto) {
	vm.Set("_toXML", func(call otto.FunctionCall) otto.Value {
		return undefined
	})
}
