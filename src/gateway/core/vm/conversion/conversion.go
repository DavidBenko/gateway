package conversion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gateway/logreport"

	"github.com/clbanning/mxj"
	"github.com/robertkrimen/otto"
)

type converter func(data interface{}) (string, error)

var undefined = otto.Value{}

func IncludeConversion(vm *otto.Otto) {
	setToJSON(vm)
	setToXML(vm)

	scripts := []string{
		// Ensure the top level AP object exists or create it
		"var AP = AP || {};",
		// Create the Conversion object
		"AP.Conversion = AP.Conversion || {};",
		"AP.Conversion.toJson = _toJson; delete _toJson;",
		"AP.Conversion.toXML = _toXML; delete _toXML;",
	}

	for _, s := range scripts {
		vm.Run(s)
	}
}

func setToJSON(vm *otto.Otto) {
	vm.Set("_toJson", func(call otto.FunctionCall) otto.Value {
		val, err := convert(vm, call, ToJSON)
		if err != nil {
			logreport.Print(err)
			return undefined
		}
		return val
	})
}

func setToXML(vm *otto.Otto) {
	vm.Set("_toXML", func(call otto.FunctionCall) otto.Value {
		val, err := convert(vm, call, ToXML)
		if err != nil {
			logreport.Print(err)
			return undefined
		}
		return val
	})
}

func convert(vm *otto.Otto, call otto.FunctionCall, fn converter) (otto.Value, error) {
	arg := call.Argument(0)
	if arg == undefined {
		return undefined, errors.New("undefined argument")
	}

	data, _ := arg.Export()

	result, err := fn(data)
	if err != nil {
		return undefined, fmt.Errorf("failed to convert: %s", err)
	}

	val, err := vm.ToValue(result)
	if err != nil {
		return undefined, fmt.Errorf("error parsing VM value: %s", err)
	}

	return val, nil
}

func ToXML(data interface{}) (string, error) {
	var json map[string]interface{}

	if j, ok := data.(map[string]interface{}); ok {
		json = j
	} else {
		return "", fmt.Errorf("invalid type: %T", data)
	}

	m := mxj.Map(json)

	buf := new(bytes.Buffer)
	err := m.XmlWriter(buf)

	if err != nil {
		return "", fmt.Errorf("error parsing JSON to XML: %s", err)
	}

	return buf.String(), nil
}

func ToJSON(data interface{}) (string, error) {
	m, err := ParseXMLToMap(data)

	if err != nil {
		return "", fmt.Errorf("errors parsing XML to JSON: %s", err)
	}

	json, err := json.Marshal(m)

	if err != nil {
		return "", fmt.Errorf("errors converting to JSON: %s", err)
	}

	return string(json), nil
}

func ParseXMLToMap(data interface{}) (*mxj.Map, error) {
	var xml string

	if x, ok := data.(string); ok {
		xml = x
	} else {
		return nil, fmt.Errorf("invalid type: %T", data)
	}

	buf := []byte(xml)

	m, err := mxj.NewMapXml(buf)
	if err != nil {
		return nil, fmt.Errorf("errors parsing XML: %s", err)
	}

	return &m, nil
}
