package model

type Route struct {
	API      *API
	Endpoint *ProxyEndpoint

	ID     int64
	Method string
	Path   string
}

type ProxyEndpoint struct {
	API         *API
	Group       *EndpointGroup
	Environment *Environment

	ID          int64
	Name        string
	Desc        string
	Active      bool
	CORSEnabled bool
	CORSAllow   string
}

type ProxyEndpointComponent struct {
	Endpoint *ProxyEndpoint

	ID          int64
	Conditional string
	Position    int64
}

type ProxyEndpointCall struct {
	Component      *ProxyEndpointComponent
	RemoteEndpoint *RemoteEndpoint

	ID          int64
	Conditional string
	Position    int64
}

type ProxyEndpointTransformationType string

const (
	TransformationJavascript ProxyEndpointTransformationType = "javascript"
)

type ProxyEndpointTransformation struct {
	Owner     interface{}
	Component *ProxyEndpointComponent
	Call      *ProxyEndpointCall

	ID       int64
	Position int64
	Before   bool
	Type     ProxyEndpointTransformationType
	Data     string
}
