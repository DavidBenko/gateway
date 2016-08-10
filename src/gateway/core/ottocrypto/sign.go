package ottocrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"gateway/crypto"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

// Default padding scheme used if nothing is supplied in the options.
var defaultPaddingScheme = "pkcs1v15"

// IncludeSigning adds the _sign function to the otto VM.
func IncludeSigning(vm *otto.Otto) {
	setSign(vm)
	setVerify(vm)
}

func setSign(vm *otto.Otto) {
	vm.Set("_sign", func(call otto.FunctionCall) otto.Value {
		d, err := getArgument(call, 0)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		data := []byte(d.(string))

		o, err := getArgument(call, 1)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		options := o.(map[string]interface{})

		var key interface{}
		if k, ok := options["key"]; ok {
			key = privateKey(k.(string))
		}

		algorithm := defaultHashAlgorithm
		if a, ok := options["algorithm"]; ok {
			algorithm = a.(string)
		}

		padding := defaultPaddingScheme
		if p, ok := options["padding"]; ok {
			padding = p.(string)
		}

		results, err := crypto.Sign(data, key, algorithm, padding)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		switch results.(type) {
		case *crypto.RsaSignature:
			r := results.(*crypto.RsaSignature)
			sr, _ := json.Marshal(r)

			return toOttoObjectValue(vm, string(sr))
		case *crypto.EcdsaSignature:
			r := results.(*crypto.EcdsaSignature)
			sr, _ := json.Marshal(r)

			return toOttoObjectValue(vm, string(sr))
		default:
			return undefined
		}
	})
}

func toOttoObjectValue(vm *otto.Otto, s string) otto.Value {
	obj, err := vm.Object(fmt.Sprintf("(%s)", string(s)))

	if err != nil {
		logreport.Print(err)
		return undefined
	}
	result, err := vm.ToValue(obj)
	if err != nil {
		logreport.Print(err)
		return undefined
	}
	return result

}

func setVerify(vm *otto.Otto) {
	vm.Set("_verify", func(call otto.FunctionCall) otto.Value {
		d, err := getArgument(call, 0)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		data := []byte(d.(string))

		signature, err := getArgument(call, 1)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		o, err := getArgument(call, 2)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		options := o.(map[string]interface{})

		var key interface{}
		if k, ok := options["key"]; ok {
			key = publicKey(k.(string))
		}

		algorithm := defaultHashAlgorithm
		if a, ok := options["algorithm"]; ok {
			algorithm = a.(string)
		}

		padding := defaultPaddingScheme
		if p, ok := options["padding"]; ok {
			padding = p.(string)
		}

		results, err := crypto.Verify(data, signature.(string), key, algorithm, padding)

		if err != nil {
			logreport.Print(err)
		}

		val, err := otto.ToValue(results)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		return val
	})
}

// All this stuff will disappear when the actual key/cert stuff is completed.
var _privateKey *rsa.PrivateKey
var generated bool

func generate() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logreport.Print(err)
	} else {
		_privateKey = key
		generated = true
	}
}

func privateKey(name string) interface{} {
	// TODO: Needs to be completed with key store, for now just generate
	// 	 a random one.
	if !generated {
		generate()
	}

	return _privateKey
}

func publicKey(name string) interface{} {
	if !generated {
		generate()
	}

	return _privateKey.Public()
}
