package conversion

import (
	"errors"
	"fmt"
	"gateway/logreport"

	"github.com/clbanning/mxj"
	"github.com/robertkrimen/otto"
)

type pather func(data interface{}, path string, subkeys []string) ([]interface{}, error)

func IncludePath(vm *otto.Otto) {
	setXMLPath(vm)
	setJSONPath(vm)
}

func setXMLPath(vm *otto.Otto) {
	vm.Set("_XMLPath", func(call otto.FunctionCall) otto.Value {
		val, err := getPath(vm, call, XMLPath)
		if err != nil {
			logreport.Print(err)
			return undefined
		}
		return val
	})
}

func setJSONPath(vm *otto.Otto) {
	vm.Set("_JSONPath", func(call otto.FunctionCall) otto.Value {
		val, err := getPath(vm, call, JSONPath)
		if err != nil {
			logreport.Print(err)
			return undefined
		}
		return val
	})
}

func getPath(vm *otto.Otto, call otto.FunctionCall, fn pather) (otto.Value, error) {
	d := call.Argument(0)
	if d == undefined {
		return undefined, errors.New("undefined data argument")
	}

	data, _ := d.Export()

	p := call.Argument(1)
	if p == undefined {
		return undefined, errors.New("undefined path argument")
	}

	pExport, _ := p.Export()

	var path string
	if strP, ok := pExport.(string); ok {
		path = strP
	} else {
		return undefined, errors.New("path should be a string")
	}

	var subkeys []string
	s := call.Argument(2)
	if s == undefined {
		subkeys = make([]string, 0)
	} else {
		export, _ := s.Export()
		if e, ok := export.([]interface{}); ok {
			subkeys = convertToStringSlice(e)
		} else {
			return undefined, errors.New("subkeys should be an array of strings")
		}
	}

	result, err := fn(data, path, subkeys)
	if err != nil {
		return undefined, fmt.Errorf("failed to convert: %s", err)
	}

	val, err := vm.ToValue(result)
	if err != nil {
		return undefined, fmt.Errorf("error parsing VM value: %s", err)
	}
	return val, nil
}

func convertToStringSlice(values []interface{}) []string {
	stringValues := make([]string, len(values))

	for i, v := range values {
		if a, ok := v.(string); ok {
			stringValues[i] = a
		}
	}

	return stringValues
}

func XMLPath(data interface{}, path string, subkeys []string) ([]interface{}, error) {
	m, err := ParseXMLToMap(data)
	if err != nil {
		return nil, err
	}

	return m.ValuesForPath(path, subkeys[:]...)
}

func JSONPath(data interface{}, path string, subkeys []string) ([]interface{}, error) {
	m := mxj.Map(data.(map[string]interface{}))
	return m.ValuesForPath(path, subkeys[:]...)
}
