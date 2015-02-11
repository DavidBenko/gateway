package proxy

import (
	"gateway/model"
	"gateway/proxy/vm"
)

func (s *Server) runJSComponent(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	var data string
	var scripts []interface{}
	for _, t := range component.BeforeTransformations {
		/* TODO: Check type, we're only working with ProxyEndpointTransformationTypeJS */
		t.Data.Unmarshal(&data)
		scripts = append(scripts, data)
	}
	component.Data.Unmarshal(&data)
	scripts = append(scripts, data)
	for _, t := range component.AfterTransformations {
		/* TODO: Check type, we're only working with ProxyEndpointTransformationTypeJS */
    t.Data.Unmarshal(&data)
		scripts = append(scripts, data)
	}

	_, err := vm.RunAll(scripts)
	return err
}

// data, _ := strings.Unquote(string(t.Data))
