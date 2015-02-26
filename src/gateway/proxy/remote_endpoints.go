package proxy

import (
	"encoding/json"
	"fmt"
	"gateway/model"
)

func (s *Server) prepareRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	switch endpoint.Type {
	case model.RemoteEndpointTypeHTTP:
		return s.prepareHTTPRequest(endpoint, data)
	}
	return nil, fmt.Errorf("%s is not a valid call type", endpoint.Type)
}

func (s *Server) prepareHTTPRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	var request HTTPRequest
	if err := json.Unmarshal(*data, &request); err != nil {
		return nil, err
	}

	var endpointData HTTPRequest
	if err := json.Unmarshal(endpoint.Data, &endpointData); err != nil {
		return nil, err
	}
	s.updateHTTPRequest(&request, &endpointData)

	if endpoint.SelectedEnvironmentData != nil {
		if err := json.Unmarshal(*endpoint.SelectedEnvironmentData, &endpointData); err != nil {
			return nil, err
		}
		s.updateHTTPRequest(&request, &endpointData)
	}

	return &request, nil
}

func (s *Server) updateHTTPRequest(request, endpointData *HTTPRequest) {
	if endpointData.Method != "" {
		request.Method = endpointData.Method
	}
	if endpointData.URL != "" {
		request.URL = endpointData.URL
	}
	if endpointData.Body != "" {
		request.Body = endpointData.Body
	}
	for name, value := range endpointData.Query {
		if request.Query == nil {
			request.Query = make(map[string]string)
		}
		request.Query[name] = value
	}
	for name, value := range endpointData.Headers {
		if request.Headers == nil {
			request.Headers = make(map[string]interface{})
		}
		request.Headers[name] = value
	}
}
