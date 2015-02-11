package proxy

import (
	"gateway/model"
	"gateway/proxy/vm"
	"strconv"
)

func (s *Server) runJSComponent(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	script, err := strconv.Unquote(string(component.Data))
	if err != nil {
		return err
	}
	_, err = vm.Run(script)
	return err
}

func (s *Server) runCallComponent(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	return nil
}
