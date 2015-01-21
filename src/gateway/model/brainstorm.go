package model

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
