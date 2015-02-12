package proxy

import (
	"gateway/model"
	"gateway/proxy/vm"
	"strconv"
)

func (s *Server) evaluateComponentConditional(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) (bool, error) {
	if component.Conditional == "" {
		return true, nil
	}

	value, err := vm.Run(component.Conditional)
	if err != nil {
		return false, err
	}

	result, err := value.ToBoolean()
	return (result == component.ConditionalPositive), err
}

func (s *Server) runJSComponentCore(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	script, err := strconv.Unquote(string(component.Data))
	if err != nil {
		return err
	}
	_, err = vm.Run(script)
	return err
}

func (s *Server) runCallComponentCore(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	return nil
}
