package request

import (
	"encoding/json"
	"fmt"
	"gateway/model"

	"github.com/y0ssar1an/q"
)

type genericKeyRequest struct {
	reqType string
}

type KeyCreateRequest struct {
	Name     string
	Contents string
}

type KeyGenerateRequest struct {
}

type KeyDeleteRequest struct {
}

type KeyResponse struct {
	Data map[string]interface{} `json:"data"`
}

func (r *KeyResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

func (r *KeyResponse) Log() string {
	return ""
}

func NewKeyRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	q.Q(data)
	return nil, nil
}

func newKeyCreateRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	return &KeyCreateRequest{}, nil
}

func (r *KeyCreateRequest) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *KeyCreateRequest) Log(devMode bool) string {
	return fmt.Sprintf("\ncreating key \"%s\"", r.Name)
}

func (r *KeyCreateRequest) Perform() Response {
	return &KeyResponse{}
}
