package crypto

import (
	"crypto/rand"
	b64 "encoding/base64"
	"gateway/logreport"

	corevm "gateway/core/vm"

	"github.com/robertkrimen/otto"
)

// IncludeRand adds the AP.Crypto.rand function to the VM.
func IncludeRand(vm *otto.Otto) {
	setRand(vm)

	scripts := []string{
		"var AP = AP || {};",
		"AP.Crypto = AP.Crypto || {};",
		"AP.Crypto.rand = _rand; delete _rand;",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}

func setRand(vm *otto.Otto) {
	vm.Set("_rand", func(call otto.FunctionCall) otto.Value {
		n, err := corevm.GetArgument(call, 0)
		if err != nil {
			logreport.Println(err)
			return otto.UndefinedValue()
		}

		number, ok := n.(int64)
		if !ok {
			logreport.Println("number of bytes should be an integer")
			return otto.UndefinedValue()
		}

		b := make([]byte, number)
		_, err = rand.Read(b)
		if err != nil {
			logreport.Println(err)
			return otto.UndefinedValue()
		}

		val, err := vm.ToValue(b64.StdEncoding.EncodeToString(b))
		if err != nil {
			logreport.Println(err)
			return otto.UndefinedValue()
		}

		return val
	})
}
