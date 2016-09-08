package ottocrypto

import (
	"errors"
	"gateway/crypto"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

var undefined = otto.Value{}

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
		password, err := getArgument(call, 0)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		iterations, err := getArgument(call, 1)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		result, err := crypto.HashPassword(password.(string), int(iterations.(int64)))

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		val, err := vm.ToValue(result)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		return val
	})
}

func setCompareHashAndPassword(vm *otto.Otto) {
	vm.Set("_compareHashAndPassword", func(call otto.FunctionCall) otto.Value {
		hash, err := getArgument(call, 0)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		password, err := getArgument(call, 1)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		result, err := crypto.CompareHashAndPassword(hash.(string), password.(string))
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		val, err := vm.ToValue(result)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		return val
	})
}

func setHash(vm *otto.Otto) {
	vm.Set("_hash", func(call otto.FunctionCall) otto.Value {
		data, err := getArgument(call, 0)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		algorithm, err := getArgument(call, 1)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		result, err := crypto.Hash(data.(string), algorithm.(string))

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		val, err := vm.ToValue(result)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		return val
	})
}

func setHashHmac(vm *otto.Otto) {
	vm.Set("_hashHmac", func(call otto.FunctionCall) otto.Value {
		data, err := getArgument(call, 0)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		tag, err := getArgument(call, 1)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		algorithm, err := getArgument(call, 2)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		result, err := crypto.HashHmac(data.(string), tag.(string), algorithm.(string))

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		val, err := vm.ToValue(result)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		return val
	})
}

func getArgument(call otto.FunctionCall, index int) (interface{}, error) {
	arg := call.Argument(index)
	if arg == undefined {
		return nil, errors.New("undefined argument")
	}

	return arg.Export()
}
