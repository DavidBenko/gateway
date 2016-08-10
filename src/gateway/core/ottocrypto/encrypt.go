package ottocrypto

import (
	"encoding/base64"
	"gateway/crypto"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

// Default hashing algorithm used if nothing is supplied in the options.
var defaultHashAlgorithm = "sha256"

func IncludeEncryption(vm *otto.Otto) {
	setEncrypt(vm)
	setDecrypt(vm)
}

func setEncrypt(vm *otto.Otto) {
	vm.Set("_encrypt", func(call otto.FunctionCall) otto.Value {
		data, err := getArgument(call, 0)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		o, err := getArgument(call, 1)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		options := o.(map[string]interface{})

		var key interface{}
		if k, ok := options["key"]; ok {
			key = publicKey(k.(string))
		}

		tag := ""
		if t, ok := options["tag"]; ok {
			tag = t.(string)
		}

		// default hashing algorithm is sha256
		algorithm := defaultHashAlgorithm
		if a, ok := options["algorithm"]; ok {
			algorithm = a.(string)
		}

		result, err := crypto.Encrypt([]byte(data.(string)), key, algorithm, tag)
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

func setDecrypt(vm *otto.Otto) {
	vm.Set("_decrypt", func(call otto.FunctionCall) otto.Value {
		d, err := getArgument(call, 0)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

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

		tag := ""
		if t, ok := options["tag"]; ok {
			tag = t.(string)
		}

		algorithm := defaultHashAlgorithm
		if a, ok := options["algorithm"]; ok {
			algorithm = a.(string)
		}

		// default expects data to be base64 encoded
		b64encoding := true
		if b, ok := options["base64"]; ok {
			b64encoding = b.(bool)
		}

		var data []byte
		if b64encoding {
			data, err = base64.StdEncoding.DecodeString(d.(string))
			if err != nil {
				logreport.Print(err)
				return undefined
			}
		} else {
			data = []byte(d.(string))
		}

		result, err := crypto.Decrypt(data, key, algorithm, tag)
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
