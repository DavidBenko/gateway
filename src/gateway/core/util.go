package core

import (
	"errors"
	"fmt"

	"gateway/core/vm"

	"github.com/robertkrimen/otto"
	"github.com/xeipuuv/gojsonschema"
)

func (c *Core) ProcessSchema(schema, body string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	bodyLoader := gojsonschema.NewStringLoader(body)
	result, err := gojsonschema.Validate(schemaLoader, bodyLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		err := ""
		for _, description := range result.Errors() {
			err += fmt.Sprintf(" - %v", description)
		}
		return errors.New(err)
	}

	return nil
}

func (c *Core) ObjectJSON(vm *vm.CoreVM, object otto.Value) (string, error) {
	jsJSON, err := vm.Object("JSON")
	if err != nil {
		return "", err
	}
	result, err := jsJSON.Call("stringify", object)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
