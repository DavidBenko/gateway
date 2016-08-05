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

// IncludeSigning adds the _sign function to the otto VM.
func IncludeSigning(vm *otto.Otto) {
	setSign(vm)
	setVerify(vm)
	setVerifyEcdsa(vm)
}

func setSign(vm *otto.Otto) {
	vm.Set("_sign", func(call otto.FunctionCall) otto.Value {
		d, err := getArgument(call, 0, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		data := []byte(d.(string))

		keyName, err := getArgument(call, 1, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		algorithm, err := getArgument(call, 2, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		key := privateKey(keyName.(string))

		padding, err := getArgument(call, 3, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		results, err := crypto.Sign(data, key, algorithm.(string), padding.(string))

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

func setVerifyEcdsa(vm *otto.Otto) {
	vm.Set("_verifyEcdsa", func(call otto.FunctionCall) otto.Value {
		/*d, err := getArgument(call, 0, ottoString)*/

		//if err != nil {
		//logreport.Print(err)
		//return undefined
		//}

		//data := []byte(d.(string))

		//s, err := getArgument(call, 1, ottoString)

		//if err != nil {
		//logreport.Print(err)
		//return undefined
		//}

		//r, err := getArgument(call, 2, ottoString)

		//if err != nil {
		//logreport.Print(err)
		//return undefined
		/*}*/

		return undefined
	})
}

func setVerify(vm *otto.Otto) {
	vm.Set("_verifyRsa", func(call otto.FunctionCall) otto.Value {
		d, err := getArgument(call, 0, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		data := []byte(d.(string))

		s, err := getArgument(call, 1, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		signature := &crypto.RsaSignature{Signature: s.(string)}

		keyName, err := getArgument(call, 2, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		algorithm, err := getArgument(call, 3, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		privateKey := privateKey(keyName.(string))
		publicKey := &privateKey.(*rsa.PrivateKey).PublicKey

		padding, err := getArgument(call, 4, ottoString)

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		results, err := crypto.Verify(data, signature, publicKey, algorithm.(string), padding.(string))

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

var _privateKey *rsa.PrivateKey
var generated bool

func privateKey(name string) interface{} {
	// TODO: Needs to be completed with key store, for now just generate
	// 	 a random one.
	if !generated {
		// generate a new one
		logreport.Println("Generating new RSA key")
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			logreport.Print(err)
			return nil
		}
		_privateKey = key
		generated = true
	}

	return _privateKey
}
