package core

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"gateway/core/vm"

	"gateway/model"
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

func (s *Core) RunComponents(vm *vm.CoreVM, components []*model.ProxyEndpointComponent) error {
	connections := make(map[int64]io.Closer)

	for _, c := range components {
		if sh := c.SharedComponentHandle; sh != nil {
			c = sh
		}
		stop, err := s.runComponent(vm, c, connections)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}

	for _, conn := range connections {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Core) runComponent(vm *vm.CoreVM, component *model.ProxyEndpointComponent, connections map[int64]io.Closer) (bool, error) {
	run, err := s.evaluateComponentConditional(vm, component)
	b := false
	if err != nil {
		return b, err
	}
	if !run {
		return b, nil
	}

	switch component.Type {
	case model.ProxyEndpointComponentTypeSingle:
		fallthrough
	case model.ProxyEndpointComponentTypeMulti:
		if err = s.runCallComponentSetup(vm, component); err != nil {
			return false, err
		}
	}

	err = s.runTransformations(vm, component.BeforeTransformations)
	if err != nil {
		return b, err
	}

	switch component.Type {
	case model.ProxyEndpointComponentTypeSingle:
		fallthrough
	case model.ProxyEndpointComponentTypeMulti:
		err = s.runCallComponentCore(vm, component, connections)
	case model.ProxyEndpointComponentTypeJS:
		b, err = s.runJSComponentCore(vm, component)
	default:
		return b, fmt.Errorf("%s is not a valid component type", component.Type)
	}
	if err != nil || b {
		return b, err
	}

	err = s.runTransformations(vm, component.AfterTransformations)
	return b, err
}

func (s *Core) runJSComponentCore(vm *vm.CoreVM, component *model.ProxyEndpointComponent) (bool, error) {
	script, err := strconv.Unquote(string(component.Data))
	if err != nil || script == "" {
		return false, err
	}
	_, stop, err := vm.RunWithStop(script)
	return stop, err
}

func (s *Core) runCallComponentSetup(vm *vm.CoreVM, component *model.ProxyEndpointComponent) error {
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

func (s *Core) runCallComponentCore(vm *vm.CoreVM, component *model.ProxyEndpointComponent, connections map[int64]io.Closer) error {
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

func (s *Core) evaluateComponentConditional(vm *vm.CoreVM, component *model.ProxyEndpointComponent) (bool, error) {
	return s.evaluateConditional(vm, component.Conditional, component.ConditionalPositive)
}

func (s *Core) evaluateCallConditional(vm *vm.CoreVM, call *model.ProxyEndpointCall) (bool, error) {
	return s.evaluateConditional(vm, call.Conditional, call.ConditionalPositive)
}

func (s *Core) evaluateConditional(vm *vm.CoreVM, conditional string, expected bool) (bool, error) {
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
