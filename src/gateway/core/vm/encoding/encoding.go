package encoding

import (
	b64 "encoding/base64"
	hex "encoding/hex"
	"errors"
	corevm "gateway/core/vm"
	"gateway/logreport"

	"github.com/robertkrimen/otto"
)

type encoder func(src []byte) string
type decoder func(s string) ([]byte, error)

var undefined = otto.Value{}

func IncludeEncoding(vm *otto.Otto) {
	setToBase64(vm)
	setToHex(vm)
	setFromBase64(vm)
	setFromHex(vm)

	scripts := []string{
		// Ensure the top level AP object exists or create it
		"var AP = AP || {};",
		// Create the Encoding object
		"AP.Encoding = AP.Encoding || {};",
		"AP.Encoding.toBase64 = _toBase64; delete _toBase64;",
		"AP.Encoding.fromBase64 = _fromBase64; delete _fromBase64;",
		"AP.Encoding.toHex = _toHex; delete _toHex;",
		"AP.Encoding.fromHex = _fromHex; delete _fromHex;",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}

func encode(vm *otto.Otto, call otto.FunctionCall, fn encoder) otto.Value {
	data, err := getDataArgAsString(call)
	if err != nil {
		logreport.Println(err)
		return undefined
	}

	encoded := fn([]byte(data))

	val, err := vm.ToValue(encoded)
	if err != nil {
		logreport.Println(err)
		return undefined
	}

	return val
}

func decode(vm *otto.Otto, call otto.FunctionCall, fn decoder) otto.Value {
	data, err := getDataArgAsString(call)
	if err != nil {
		logreport.Println(err)
		return undefined
	}

	d, err := fn(data)
	if err != nil {
		logreport.Println(err)
		return undefined
	}

	val, err := vm.ToValue(string(d[:]))
	if err != nil {
		logreport.Println(err)
		return undefined
	}

	return val
}

func getDataArgAsString(call otto.FunctionCall) (string, error) {
	d, err := corevm.GetArgument(call, 0)
	if err != nil {
		return "", err
	}

	if ds, ok := d.(string); ok {
		return ds, nil
	}
	return "", errors.New("data should be a string")
}

func setToBase64(vm *otto.Otto) {
	vm.Set("_toBase64", func(call otto.FunctionCall) otto.Value {
		fn := b64.StdEncoding.EncodeToString
		return encode(vm, call, fn)
	})
}

func setFromBase64(vm *otto.Otto) {
	vm.Set("_fromBase64", func(call otto.FunctionCall) otto.Value {
		fn := b64.StdEncoding.DecodeString
		return decode(vm, call, fn)
	})
}

func setToHex(vm *otto.Otto) {
	vm.Set("_toHex", func(call otto.FunctionCall) otto.Value {
		fn := hex.EncodeToString
		return encode(vm, call, fn)
	})
}

func setFromHex(vm *otto.Otto) {
	vm.Set("_fromHex", func(call otto.FunctionCall) otto.Value {
		fn := hex.DecodeString
		return decode(vm, call, fn)
	})
}
