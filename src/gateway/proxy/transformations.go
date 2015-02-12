package proxy

import (
	"gateway/model"
	"gateway/proxy/vm"
	"strconv"
)

func (s *Server) runJSTransformation(vm *vm.ProxyVM,
	transformation *model.ProxyEndpointTransformation) error {

	script, err := strconv.Unquote(string(transformation.Data))
	if err != nil || script == "" {
		return err
	}
	_, err = vm.Run(script)
	return err
}
