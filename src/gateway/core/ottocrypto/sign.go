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
			logreport.Println(err)
			return undefined
		}

		data, ok := d.(string)
		if !ok {
			logreport.Println("data should be a string")
			return undefined
		}

		o, err := getArgument(call, 1)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		options, ok := o.(map[string]interface{})
		if !ok {
			logreport.Println("options should be an object")
			return undefined
		}

		key, algorithm, _, err := getOptions(o, keySource, accountID)

		padding, err := getOptionString(options, "padding", true)
		if err != nil {
			logreport.Println(err)
			return undefined
		}
		if padding == "" {
			padding = DefaultPaddingScheme
		}

		results, err := crypto.Sign([]byte(data), key, algorithm, padding)

		if err != nil {
			logreport.Println(err)
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
			logreport.Println(err)
			return undefined
		}

		data, ok := d.(string)
		if !ok {
			logreport.Println("data should be a string")
			return undefined
		}

		s, err := getArgument(call, 1)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		signature, ok := s.(string)
		if !ok {
			logreport.Println("signature should be a string")
			return undefined
		}

		o, err := getArgument(call, 2)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		options, ok := o.(map[string]interface{})
		if !ok {
			logreport.Println("options should be an object")
			return undefined
		}

		key, algorithm, _, err := getOptions(o, keySource, accountID)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		padding, err := getOptionString(options, "padding", true)
		if err != nil {
			logreport.Println(err)
			return undefined
		}
		if padding == "" {
			padding = DefaultPaddingScheme
		}

		results, err := crypto.Verify([]byte(data), signature, key, algorithm, padding)

		if err != nil {
			logreport.Println(err)
		}

		val, err := otto.ToValue(results)

		if err != nil {
			logreport.Println(err)
			return undefined
		}

		return val
	})
}
