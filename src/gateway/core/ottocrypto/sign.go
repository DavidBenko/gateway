package ottocrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"gateway/crypto"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

// IncludeSigning adds the _sign function to the otto VM.
func IncludeSigning(vm *otto.Otto) {
	setSign(vm)
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

		results, err := crypto.Sign(data, key, algorithm.(string))

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		val, err := otto.ToValue(base64.StdEncoding.EncodeToString(results))

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		return val
	})
}

func privateKey(name string) interface{} {
	// TODO: Needs to be completed with key store, for now just generate
	// 	 a random one.
	key, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		logreport.Print(err)
		return nil
	}

	return key
}
