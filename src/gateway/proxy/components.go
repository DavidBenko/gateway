package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/robertkrimen/otto"

	"gateway/model"
	"gateway/proxy/vm"
)

/**
 * TODO: Hand error control back to clients where appropriate.
 * If an http proxy request fails, do we shit ourselves, or give the user a
 * chance to handle it in a custom manner? I think for many answers we do the
 * latter, so we need to turn some errors into values that can be checked
 * and called.
 */

/**
* TODO: Desperately need a debug mode to output more helpful errors.
* Without a wrapper error and custom messaging, bubbling up root err is not
* that helpful.
 */

func (s *Server) runComponents(vm *vm.ProxyVM, components []*model.ProxyEndpointComponent) error {
	connections := make(map[int64]io.Closer)

	for _, c := range components {
		if sh := c.SharedComponentHandle; sh != nil {
			c = sh
		}
		if err := s.runComponent(vm, c, connections); err != nil {
			return err
		}
	}

	for _, conn := range connections {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) runComponent(vm *vm.ProxyVM, component *model.ProxyEndpointComponent, connections map[int64]io.Closer) error {
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
		err = s.runCallComponentCore(vm, component, connections)
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

	return nil
}

func (s *Server) runJSComponentCore(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	script, err := strconv.Unquote(string(component.Data))
	if err != nil || script == "" {
		return err
	}
	wrappedScript, getter := wrapJSComponent(vm, script)
	_, err = vm.Run(wrappedScript)
	if err != nil {
		return err
	}
	return getter()
}

func wrapJSComponent(vm *vm.ProxyVM, script string) (string, func() error) {
	breakVar := "__vm_break_value"
	hash := "8a52973428f63bb0135a3abf535fec0f15b4c8eda1e9a2f1431f0a1f759babd3"
	vm.Run(fmt.Sprintf("const break = '%s';", hash))

	wrapped := fmt.Sprintf("var %s = (function() {%s})();", breakVar, script)

	undefined := otto.Value{}
	fn := func() error {
		v, err := vm.Get(breakVar)
		if err != nil {
			return err
		}
		val, err := v.Export()
		if err != nil {
			return err
		}
		if v == undefined {
			return nil
		}
		e, err := json.Marshal(val)
		if err != nil {
			return err
		}
		return errors.New(string(e[:]))
	}

	return wrapped, fn
}

func (s *Server) runCallComponentSetup(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) error {
	script := ""
	for _, c := range component.AllCalls() {
		name, err := c.Name()
		if err != nil {
			return err
		}
		script = script + fmt.Sprintf("var %s = %s || new AP.Call();\n", name, name)
	}

	_, err := vm.Run(script)
	return err
}

func (s *Server) runCallComponentCore(vm *vm.ProxyVM, component *model.ProxyEndpointComponent, connections map[int64]io.Closer) error {
	var activeCalls []*model.ProxyEndpointCall

	for _, call := range component.AllCalls() {
		run, err := s.evaluateCallConditional(vm, call)
		if err != nil {
			return err
		}
		if run {
			// We don't want to litter the VM with before transformation code
			// while we're still evaluating conditionals
			activeCalls = append(activeCalls, call)
		}
	}

	var activeCallNames []string
	for _, call := range activeCalls {
		err := s.runTransformations(vm, call.BeforeTransformations)
		if err != nil {
			return err
		}

		name, err := call.Name()
		if err != nil {
			return err
		}
		activeCallNames = append(activeCallNames, name)
	}

	requests, err := s.getRequests(vm, activeCallNames, activeCalls, connections)
	if err != nil {
		return err
	}

	responses, err := s.makeRequests(vm, requests)
	if err != nil {
		return err
	}

	responsesJSON, err := json.Marshal(responses)
	if err != nil {
		return err
	}

	// TODO(binary132): move this into "gateway/proxy/vm" calls.go
	responsesScript := fmt.Sprintf("AP.insertResponses([%s],%s);",
		strings.Join(activeCallNames, ","), responsesJSON)
	_, err = vm.Run(responsesScript)
	if err != nil {
		return err
	}

	for _, call := range activeCalls {
		err := s.runTransformations(vm, call.AfterTransformations)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) evaluateComponentConditional(vm *vm.ProxyVM, component *model.ProxyEndpointComponent) (bool, error) {
	return s.evaluateConditional(vm, component.Conditional, component.ConditionalPositive)
}

func (s *Server) evaluateCallConditional(vm *vm.ProxyVM, call *model.ProxyEndpointCall) (bool, error) {
	return s.evaluateConditional(vm, call.Conditional, call.ConditionalPositive)
}

func (s *Server) evaluateConditional(vm *vm.ProxyVM, conditional string, expected bool) (bool, error) {
	trimmedConditional := strings.TrimSpace(conditional)
	if trimmedConditional == "" {
		return true, nil
	}

	value, err := vm.Run(trimmedConditional)
	if err != nil {
		return false, err
	}

	result, err := value.ToBoolean()
	return (result == expected), err
}
