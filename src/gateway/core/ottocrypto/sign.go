package ottocrypto

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

func IncludeSigning(vm *otto.Otto) {
	setSign(vm)
}

func setSign(vm *otto.Otto) {
	vm.Set("_sign", func(call otto.FunctionCall) otto.Value {
		undefined := otto.Value{}

		dataArg := call.Argument(0)

		if dataArg == undefined {
			logreport.Print("data is undefined")
			return undefined
		}

		d, err := dataArg.ToString()

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		data := []byte(d)

		keyNameArg := call.Argument(1)

		if keyNameArg == undefined {
			logreport.Print("keyName is undefined")
			return undefined
		}

		keyName, err := keyNameArg.ToString()

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		algoArg := call.Argument(2)

		if algoArg == undefined {
			logreport.Print("algorithm is undefined")
			return undefined
		}

		algorithm, err := algoArg.ToString()

		if err != nil {
			logreport.Print(err)
			return undefined
		}

		key := privateKey(keyName)

		results, err := sign(data, key, algorithm)

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
	key, err := rsa.GenerateKey(rand.Reader, 256)

	if err != nil {
		logreport.Print(err)
		return nil
	}

	return key
}

func sign(data []byte, privKey interface{}, algorithm string) ([]byte, error) {
	switch privKey.(type) {
	case *rsa.PrivateKey:
		logreport.Println("\nThis is an RSA key")
	case *ecdsa.PrivateKey:
		logreport.Println("\nThis is an ECDSA key")
	case *dsa.PrivateKey:
		logreport.Println("\nThis is a DSA key")
	default:
		logreport.Println("I have no idea what this is")
	}

	return nil, nil
}
