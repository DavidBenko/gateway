package ottocrypto

import (
	b64 "encoding/base64"
	"errors"
	"gateway/crypto"
	"gateway/logreport"
	"strings"

	"github.com/robertkrimen/otto"
)

func IncludeAes(vm *otto.Otto) {
	setAesEncrypt(vm)
	setAesDecrypt(vm)

	scripts := []string{
		"AP.Crypto.Aes.encrypt = _aesEncrypt; delete _aesEncrypt;",
		"AP.Crypto.Aes.decrypt = _aesDecrypt; delete _aesDecrypt;",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}

func setAesEncrypt(vm *otto.Otto) {
	vm.Set("_aesEncrypt", func(call otto.FunctionCall) otto.Value {
		k, data, err := getAesParams(call)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		result, err := crypto.EncryptAes([]byte(data), k)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		encoded := b64.StdEncoding.EncodeToString(result)

		val, err := vm.ToValue(encoded)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		return val
	})
}

func setAesDecrypt(vm *otto.Otto) {
	vm.Set("_aesDecrypt", func(call otto.FunctionCall) otto.Value {
		k, data, err := getAesParams(call)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		decoded, err := b64.StdEncoding.DecodeString(data)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		result, err := crypto.DecryptAes([]byte(decoded), k)
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		val, err := vm.ToValue(string(result))
		if err != nil {
			logreport.Println(err)
			return undefined
		}

		return val
	})
}

func getAesParams(call otto.FunctionCall) (*crypto.SymmetricKey, string, error) {
	data, err := getData(call)
	if err != nil {
		return nil, "", err
	}

	o, err := getArgument(call, 1)
	if err != nil {
		return nil, "", err
	}

	var mode crypto.AesMode
	mode = crypto.CFBMode
	options, ok := o.(map[string]interface{})
	if !ok {
		return nil, "", errors.New("options should be an object")
	}

	k, err := getOptionString(options, "key", false)
	if err != nil {
		return nil, "", err
	}

	i, _ := getOptionString(options, "iv", true)

	m, _ := getOptionString(options, "mode", true)
	switch m {
	case "cbc":
		mode = crypto.CBCMode
	case "cfb":
		mode = crypto.CFBMode
	default:
		mode = crypto.CFBMode
	}

	if strings.TrimSpace(k) == "" {
		return nil, "", errors.New("missing key")
	}

	key, err := b64.StdEncoding.DecodeString(k)
	if err != nil {
		return nil, "", err
	}

	iv, err := b64.StdEncoding.DecodeString(i)
	if err != nil {
		return nil, "", err
	}

	symkey, err := crypto.ParseAesKey(key, iv, mode)
	if err != nil {
		return nil, "", err
	}

	return symkey, data, nil
}
