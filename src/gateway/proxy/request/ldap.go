package request

import (
	"encoding/json"
	"fmt"

	aperrors "gateway/errors"
	"gateway/model"

	"github.com/go-ldap/ldap"
)

// LDAPRequest TODO
type LDAPRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LDAPResponse TODO
type LDAPResponse struct {
}

// NewLDAPRequest TODO
func NewLDAPRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := new(LDAPRequest)

	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	endpointData := new(LDAPRequest)
	if err := json.Unmarshal(endpoint.Data, endpointData); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal endpoint data: %v", err)
	}
	request.updateWith(endpointData)

	if endpoint.SelectedEnvironmentData != nil {
		if err := json.Unmarshal(*endpoint.SelectedEnvironmentData, endpointData); err != nil {
			return nil, err
		}
		request.updateWith(endpointData)
	}

	return request, nil
}

func (ldapRequest *LDAPRequest) updateWith(other *LDAPRequest) {
	if other.Username != "" {
		ldapRequest.Username = other.Username
	}

	if other.Password != "" {
		ldapRequest.Password = other.Password
	}
}

// Log TODO
func (ldapRequest *LDAPRequest) Log(devMode bool) string {
	return "TODO"
}

// Perform TODO
func (ldapRequest *LDAPRequest) Perform() Response {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", ldapRequest.Host, ldapRequest.Port))
	if err != nil {
		return NewErrorResponse(aperrors.NewWrapped("[ldap] Dialing ldap endpoint", err))
	}

	defer l.Close()

	return new(LDAPResponse)
}

// JSON TODO
func (r *LDAPResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

// Log TODO
func (r *LDAPResponse) Log() string {
	// TODO
	return "Response TODO"
}
