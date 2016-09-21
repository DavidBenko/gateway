package ottocrypto

import (
	"encoding/json"
	"gateway/crypto"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

// IncludeSigning create the AP.Crypto.sign and AP.Crypto.verify helper functions in the
// supplied Otto VM.
func IncludeSigning(vm *otto.Otto, accountID int64, keySource KeyDataSource) {
	setSign(vm, accountID, keySource)
	setVerify(vm, accountID, keySource)

	scripts := []string{
		"AP.Crypto.sign = _sign; delete _sign;",
		"AP.Crypto.verify = _verify; delete _verify;",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}

func setSign(vm *otto.Otto, accountID int64, keySource KeyDataSource) {
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
		if keyName, ok := options["key"]; ok {
			if k, found := keySource.GetKey(accountID, keyName.(string)); found {
				key = k
			}
		}

		algorithm := DefaultHashAlgorithm
		if a, ok := options["algorithm"]; ok {
			algorithm = a.(string)
		}

		padding := DefaultPaddingScheme
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

func setVerify(vm *otto.Otto, accountID int64, keySource KeyDataSource) {
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
		if keyName, ok := options["key"]; ok {
			if k, found := keySource.GetKey(accountID, keyName.(string)); found {
				key = k
			}
		}

		algorithm := DefaultHashAlgorithm
		if a, ok := options["algorithm"]; ok {
			algorithm = a.(string)
		}

		padding := DefaultPaddingScheme
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
