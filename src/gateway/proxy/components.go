package proxy

import (
	"fmt"
	"gateway/model"
	"gateway/proxy/vm"
	"strconv"
)

func (s *Server) runComponents(vm *vm.ProxyVM, components []*model.ProxyEndpointComponent) error {
	for _, c := range components {
		if err := s.runComponent(vm, c); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) runComponent(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	run, err := s.evaluateComponentConditional(vm, component)
	if err != nil {
		return err
	}
	if !run {
		return nil
	}

	switch component.Type {
	case model.ProxyEndpointComponentTypeSingle:
		fallthrough
	case model.ProxyEndpointComponentTypeMulti:
		if err = s.runCallComponentSetup(vm, component); err != nil {
			return err
		}
	}

	err = s.runTransformations(vm, component.BeforeTransformations)
	if err != nil {
		return err
	}

	switch component.Type {
	case model.ProxyEndpointComponentTypeSingle:
		fallthrough
	case model.ProxyEndpointComponentTypeMulti:
		err = s.runCallComponentCore(vm, component)
	case model.ProxyEndpointComponentTypeJS:
		err = s.runJSComponentCore(vm, component)
	default:
		return fmt.Errorf("%s is not a valid component type", component.Type)
	}
	if err != nil {
		return err
	}

	err = s.runTransformations(vm, component.AfterTransformations)
	if err != nil {
		return err
	}

	switch component.Type {
	case model.ProxyEndpointComponentTypeSingle:
		fallthrough
	case model.ProxyEndpointComponentTypeMulti:
		if err = s.runCallComponentFinalize(vm, component); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) runJSComponentCore(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	script, err := strconv.Unquote(string(component.Data))
	if err != nil || script == "" {
		return err
	}
	_, err = vm.Run(script)
	return err
}

func (s *Server) runCallComponentSetup(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	script := ""
	for _, c := range component.AllCalls() {
		name := c.EndpointNameOverride
		if name == "" {
			name = c.RemoteEndpoint.Name
		}
		script = script + fmt.Sprintf("var %s = %s || new AP.Call();\n", name, name)
	}

	_, err := vm.Run(script)
	return err
}

func (s *Server) runCallComponentCore(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	var activeCalls []*model.ProxyEndpointCall

	for _, call := range component.AllCalls() {
		run, err := s.evaluateCallConditional(vm, call)
		if err != nil {
			return err
		}
		if run {
			activeCalls = append(activeCalls, call)
		}
	}

	for _, call := range activeCalls {
		err := s.runTransformations(vm, call.BeforeTransformations)
		if err != nil {
			return err
		}
	}

	/* TODO: Prep & extract requests */
	/* TODO: Make requests */
	/* TODO: Insert responses */

	for _, call := range activeCalls {
		err := s.runTransformations(vm, call.AfterTransformations)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) runCallComponentFinalize(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	return nil
}

func (s *Server) evaluateComponentConditional(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) (bool, error) {
	return s.evaluateConditional(vm, component.Conditional, component.ConditionalPositive)
}

func (s *Server) evaluateCallConditional(vm *vm.ProxyVM, call *model.ProxyEndpointCall) (bool, error) {
	return s.evaluateConditional(vm, call.Conditional, call.ConditionalPositive)
}

func (s *Server) evaluateConditional(vm *vm.ProxyVM, conditional string, expected bool) (bool, error) {
	if conditional == "" {
		return true, nil
	}

	value, err := vm.Run(conditional)
	if err != nil {
		return false, err
	}

	result, err := value.ToBoolean()
	return (result == expected), err
}
