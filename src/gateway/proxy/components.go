package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/model"
	"gateway/proxy/vm"
	"strconv"
	"strings"
	"time"
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
		name, err := c.Name()
		if err != nil {
			return err
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

	requestScript := fmt.Sprintf("AP.prepareRequests(%s);",
		strings.Join(activeCallNames, ","))
	requestsObject, err := vm.Run(requestScript)
	if err != nil {
		return err
	}
	requestsJSON := requestsObject.String()

	var requestData []*json.RawMessage
	err = json.Unmarshal([]byte(requestsJSON), &requestData)
	if err != nil {
		return err
	}

	var abstractedRequests []Request
	for i, call := range activeCalls {
		if call.RemoteEndpoint == nil {
			return errors.New("Remote endpoint is not loaded")
		}
		request, err := s.prepareRequest(call.RemoteEndpoint, requestData[i])
		if err != nil {
			return err
		}
		abstractedRequests = append(abstractedRequests, request)
	}

	responses, err := s.makeRequests(vm, abstractedRequests)
	if err != nil {
		return err
	}
	responsesJSON, err := json.Marshal(responses)
	if err != nil {
		return err
	}
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

func (s *Server) makeRequests(vm *vm.ProxyVM, proxyRequests []Request) ([]Response, error) {
	start := time.Now()
	defer func() {
		vm.ProxiedRequestsDuration += time.Since(start)
	}()
	return s.MakeRequests(proxyRequests, vm.LogPrefix)
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
