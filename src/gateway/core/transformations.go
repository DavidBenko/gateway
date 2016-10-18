package core

import (
	"fmt"
	"strconv"

	"gateway/core/vm"
	"gateway/model"

	"github.com/jmoiron/sqlx/types"
)

func (s *Core) runTransformations(vm *vm.CoreVM,
	transformations []*model.ProxyEndpointTransformation) error {

	for _, t := range transformations {
		switch t.Type {
		case model.ProxyEndpointTransformationTypeJS:
			if err := s.runJSTransformation(vm, t); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s is not a valid transformation type", t.Type)
		}
	}

	return nil
}

func (s *Core) runJSTransformation(vm *vm.CoreVM,
	transformation *model.ProxyEndpointTransformation) error {
	return s.runStoredJSONScript(vm, transformation.Data)
}

func (s *Core) runStoredJSONScript(vm *vm.CoreVM, jsonScript types.JsonText) error {
	script, err := strconv.Unquote(string(jsonScript))
	if err != nil || script == "" {
		return err
	}
	_, err = vm.Run(script)
	return err
}
