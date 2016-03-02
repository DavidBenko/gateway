package proxy

import (
	"errors"
	"time"

	"gateway/core/request"
	"gateway/model"
	"gateway/proxy/vm"
)

// getRequests takes a slice of call names and a slice of
// *model.ProxyEndpointCalls, and gets a slice of request.Request.
func (s *Server) getRequests(
	vm *vm.ProxyVM,
	callNames []string,
	endpointCalls []*model.ProxyEndpointCall,
) ([]request.Request, error) {
	rawRequests, err := vm.PrepareRawRequests(callNames)
	if err != nil {
		return nil, err
	}

	var requests []request.Request
	for i, call := range endpointCalls {
		if call.RemoteEndpoint == nil {
			return nil, errors.New("Remote endpoint is not loaded")
		}
		request, err := s.PrepareRequest(call.RemoteEndpoint, rawRequests[i])
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

type responsePayload struct {
	index    int
	response request.Response
}

func (s *Server) makeRequests(vm *vm.ProxyVM, proxyRequests []request.Request) ([]request.Response, error) {
	start := time.Now()
	defer func() {
		vm.ProxiedRequestsDuration += time.Since(start)
	}()

	n := len(proxyRequests)
	requestDurations := make([]time.Duration, n)
	responses := make([]request.Response, n)
	c := make(chan *responsePayload, n)

	for i, req := range proxyRequests {
		go func(j int, r request.Request) { c <- &responsePayload{j, r.Perform()} }(i, req)
	}

	// TODO(binary132): parallel deserialize?
	for i := 0; i < n; i++ {
		select {
		case r := <-c:
			requestDurations[r.index] = time.Since(start)
			responses[r.index] = r.response
		}
	}

	for i, req := range proxyRequests {
		vm.LogPrint("%s [request] %s %s (%v)", vm.LogPrefix,
			req.Log(s.devMode), responses[i].Log(), requestDurations[i])
	}

	return responses, nil
}
