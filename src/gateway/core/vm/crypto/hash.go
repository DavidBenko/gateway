package crypto

import (
	"gateway/crypto"
	"gateway/logreport"

	corevm "gateway/core/vm"

	"github.com/robertkrimen/otto"
)

// IncludeHashing creates the _hashPassword, _compareHashAndPassword, hash and
// hashHmac functions in the supplied Otto VM.
func IncludeHashing(vm *otto.Otto) {
	setHashPassword(vm)
	setCompareHashAndPassword(vm)
	setHash(vm)
	setHashHmac(vm)
}

func setHashPassword(vm *otto.Otto) {
	vm.Set("_hashPassword", func(call otto.FunctionCall) otto.Value {
		p, err := corevm.GetArgument(call, 0)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		password, ok := p.(string)
		if !ok {
			logreport.Println("password should be a string")
			return undefined
		}

		i, err := corevm.GetArgument(call, 1)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		iterations, ok := i.(int64)
		if !ok {
			logreport.Println("iterations should be a number")
			return undefined
		}

		result, err := crypto.HashPassword(password, int(iterations))

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		val, err := vm.ToValue(result)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		return val
	})
}

func setCompareHashAndPassword(vm *otto.Otto) {
	vm.Set("_compareHashAndPassword", func(call otto.FunctionCall) otto.Value {
		h, err := corevm.GetArgument(call, 0)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		hash, ok := h.(string)
		if !ok {
			logreport.Println("hash should be a string")
			return undefined
		}

		p, err := corevm.GetArgument(call, 1)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		password, ok := p.(string)
		if !ok {
			logreport.Println("password should be a string")
			return undefined
		}

		result, err := crypto.CompareHashAndPassword(hash, password)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		val, err := vm.ToValue(result)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		return val
	})
}

func setHash(vm *otto.Otto) {
	vm.Set("_hash", func(call otto.FunctionCall) otto.Value {
		d, err := corevm.GetArgument(call, 0)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		data, ok := d.(string)
		if !ok {
			logreport.Println("data should be a string")
			return undefined
		}

		a, err := corevm.GetArgument(call, 1)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		algorithm, ok := a.(string)
		if !ok {
			logreport.Println("algorithm should be a string")
			return undefined
		}

		result, err := crypto.Hash(data, algorithm)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		val, err := vm.ToValue(result)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		return val
	})
}

func setHashHmac(vm *otto.Otto) {
	vm.Set("_hashHmac", func(call otto.FunctionCall) otto.Value {
		d, err := corevm.GetArgument(call, 0)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		data, ok := d.(string)
		if !ok {
			logreport.Println("data should be a string")
			return undefined
		}

		t, err := corevm.GetArgument(call, 1)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		tag, ok := t.(string)
		if !ok {
			logreport.Println("tag should be a string")
			return undefined
		}

		a, err := corevm.GetArgument(call, 2)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		algorithm, ok := a.(string)
		if !ok {
			logreport.Println("algorithm should be a string")
			return undefined
		}

		result, err := crypto.HashHmac(data, tag, algorithm)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		val, err := vm.ToValue(result)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		return val
	})
}
