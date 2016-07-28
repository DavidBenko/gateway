package ottocrypto

import (
	"encoding/base64"
	"errors"
	"gateway/crypto"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

var undefined = otto.Value{}

type OttoValueType int

const (
	ottoString = iota
	ottoInteger
)

// IncludeHashing adds the _hash, _hashPassword & _hashHmac functions to the otto VM.
func IncludeHashing(vm *otto.Otto) {
	setHashPassword(vm)
	setHash(vm)
	setHashHmac(vm)
}

func setHashPassword(vm *otto.Otto) {
	vm.Set("_hashPassword", func(call otto.FunctionCall) otto.Value {
		password, err := getArgument(call, 0, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		iterations, err := getArgument(call, 1, ottoInteger)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		result, err := crypto.HashPassword([]byte(password.(string)), int(iterations.(int64)))

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		val, err := vm.ToValue(base64.StdEncoding.EncodeToString(result))

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		return val
	})
}

func setHash(vm *otto.Otto) {
	vm.Set("_hash", func(call otto.FunctionCall) otto.Value {
		data, err := getArgument(call, 0, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		algorithm, err := getArgument(call, 1, ottoString)

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
		data, err := getArgument(call, 0, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		tag, err := getArgument(call, 1, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		algorithm, err := getArgument(call, 2, ottoString)

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

func getArgument(call otto.FunctionCall, index int, t OttoValueType) (interface{}, error) {
	arg := call.Argument(index)
	if arg == undefined {
		return nil, errors.New("undefined argument")
	}

	switch t {
	case ottoString:
		v, err := arg.ToString()
		if err != nil {
			return nil, err
		}
		return v, nil
	case ottoInteger:
		v, err := arg.ToInteger()
		if err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, errors.New("unknown otto value type")
	}
}
