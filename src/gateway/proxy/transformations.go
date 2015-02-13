package proxy

import (
	"fmt"
	"gateway/model"
	"gateway/proxy/vm"
)

func (s *Server) runTransformations(vm *vm.ProxyVM,
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

func (s *Server) runJSTransformation(vm *vm.ProxyVM,
	transformation *model.ProxyEndpointTransformation) error {
	return s.runStoredJSONScript(vm, transformation.Data)
}
