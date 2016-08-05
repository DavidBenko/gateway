package ottocrypto

import "github.com/robertkrimen/otto"

func IncludeEncryption(vm *otto.Otto) {

}

func setEncrypt(vm *otto.Otto) {
	vm.Set("_encrypt", func(call otto.FunctionCall) otto.Value {
		return undefined
	})
}
