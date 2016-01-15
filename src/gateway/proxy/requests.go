package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	aperrors "gateway/errors"
	"gateway/model"
	"gateway/proxy/request"
	"gateway/proxy/vm"
)

// getRequests takes a slice of call names and a slice of
// *model.ProxyEndpointCalls, and gets a slice of request.Request.
func (s *Server) getRequests(
	vm *vm.ProxyVM,
	callNames []string,
	endpointCalls []*model.ProxyEndpointCall,
	connections map[int64]io.Closer,
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
		request, err := s.prepareRequest(call.RemoteEndpoint, rawRequests[i], connections)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func (s *Server) prepareRequest(
	endpoint *model.RemoteEndpoint,
	data *json.RawMessage,
	connections map[int64]io.Closer,
) (request.Request, error) {
	if !model.IsRemoteEndpointTypeEnabled(endpoint.Type) {
		return nil, fmt.Errorf("Remote endpoint type %s is not enabled", endpoint.Type)
	}

	var r request.Request
	var e error

	switch endpoint.Type {
	case model.RemoteEndpointTypeHTTP:
		r, e = request.NewHTTPRequest(s.httpClient, endpoint, data)
	case model.RemoteEndpointTypeSQLServer:
		r, e = request.NewSQLServerRequest(s.dbPools, endpoint, data)
	case model.RemoteEndpointTypePostgres:
		r, e = request.NewPostgresRequest(s.dbPools, endpoint, data)
	case model.RemoteEndpointTypeMySQL:
		r, e = request.NewMySQLRequest(s.dbPools, endpoint, data)
	case model.RemoteEndpointTypeMongo:
		r, e = request.NewMongoRequest(s.dbPools, endpoint, data)
	case model.RemoteEndpointTypeSoap:
		r, e = request.NewSoapRequest(endpoint, data, s.soapConf, s.ownDb)
	case model.RemoteEndpointTypeScript:
		r, e = request.NewScriptRequest(endpoint, data)
	case model.RemoteEndpointTypeLDAP:
		r, e = request.NewLDAPRequest(endpoint, data)
	default:
		return nil, fmt.Errorf("%q is not a valid endpoint type", endpoint.Type)
	}

	if e != nil {
		return r, e
	}

	if sc, ok := r.(request.ReusableConnection); ok {
		conn, err := sc.CreateOrReuse(connections[endpoint.ID])
		if err != nil {
			return nil, aperrors.NewWrapped("[requests.go] initializing sticky connection", err)
		}
		connections[endpoint.ID] = conn
	}

	return r, nil
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
