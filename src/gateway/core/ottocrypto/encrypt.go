package ottocrypto

import (
	"encoding/base64"
	"gateway/crypto"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

type KeyDataSource interface {
	GetKey(int64, string) (interface{}, bool)
}

// Default hashing algorithm used if nothing is supplied in the options.
var defaultHashAlgorithm = "sha256"

// IncludeEncryption adds the AP.Crypto.encrypt and AP.Crypto.decrypt functions in
// the supplied Otto VM.
func IncludeEncryption(vm *otto.Otto, accountID int64, keySource KeyDataSource) {
	setEncrypt(vm, accountID, keySource)
	setDecrypt(vm, accountID, keySource)

	scripts := []string{
		"AP.Crypto.encrypt = _encrypt; delete _encrypt;",
		"AP.Crypto.decrypt = _decrypt; delete _decrypt;",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}

func setEncrypt(vm *otto.Otto, accountID int64, keySource KeyDataSource) {
	vm.Set("_encrypt", func(call otto.FunctionCall) otto.Value {
		d, err := getArgument(call, 0)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		var data string
		if ds, ok := d.(string); ok {
			data = ds
		}

		o, err := getArgument(call, 1)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		options := o.(map[string]interface{})

		key, err := GetKeyFromSource(options, keySource, accountID)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		tag, err := GetOptionString(options, "tag", true)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		// default hashing algorithm is sha256
		algorithm := defaultHashAlgorithm
		a, err := GetOptionString(options, "algorithm", true)
		if err != nil {
			logreport.Println(err)
			return undefined
		} else {
			algorithm = a
		}

		result, err := crypto.Encrypt([]byte(data), key, algorithm, tag)
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

func setDecrypt(vm *otto.Otto, accountID int64, keySource KeyDataSource) {
	vm.Set("_decrypt", func(call otto.FunctionCall) otto.Value {
		d, err := getArgument(call, 0)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		if ds, ok := d.(string); !ok {
			logreport.Println("data should be a string")
			return undefined
		}

		o, err := getArgument(call, 1)
		if err != nil {
			logreport.Print(err)
			return undefined
		}

		options := o.(map[string]interface{})

		key, err := GetKeyFromSource(options, keySource, accountID)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		tag, err := GetOptionString(options, "tag", true)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		// default hashing algorithm is sha256
		algorithm := defaultHashAlgorithm
		a, err := GetOptionString(options, "algorithm", true)
		if err != nil {
			logreport.Println(err)
			return undefined
		} else {
			algorithm = a
		}

		// default expects data to be base64 encoded
		b64encoding := true
		if b, ok := options["base64"]; ok {
			if v, ok := b.(bool); ok {
				b64encoding = v
			}
		}

		var data []byte
		if b64encoding {
			data, err = base64.StdEncoding.DecodeString(ds)
			if err != nil {
				logreport.Print(err)
				return undefined
			}
		} else {
			data = []byte(ds)
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
