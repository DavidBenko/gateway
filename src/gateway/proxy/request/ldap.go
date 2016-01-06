package request

import (
	"encoding/json"
	"fmt"

	aperrors "gateway/errors"
	"gateway/model"
	apldap "gateway/proxy/request/ldap"

	"github.com/go-ldap/ldap"
)

// LDAPOperation TODO
type LDAPOperation interface {
	Invoke(*ldap.Conn) (*apldap.Response, error)
}

// LDAPRequest TODO
type LDAPRequest struct {
	Host     string
	Port     int
	Username string
	Password string

	operationName string
	arguments     LDAPOperation
}

// UnmarshalJSON TODO
func (l *LDAPRequest) UnmarshalJSON(data []byte) error {

	if l == nil {
		return nil
	}

	var fields map[string]*json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	var arguments json.RawMessage
	for k, v := range fields {
		switch k {
		case "host":
			var host string
			if err := json.Unmarshal([]byte(*v), &host); err != nil {
				return err
			}
			l.Host = host
		case "port":
			var port int
			if err := json.Unmarshal([]byte(*v), &port); err != nil {
				return err
			}
			l.Port = port
		case "username":
			var username string
			if err := json.Unmarshal([]byte(*v), &username); err != nil {
				return err
			}
			l.Username = username
		case "password":
			var password string
			if err := json.Unmarshal([]byte(*v), &password); err != nil {
				return err
			}
			l.Password = password
		case "operationName":
			var opName string
			if err := json.Unmarshal([]byte(*v), &opName); err != nil {
				return err
			}
			l.operationName = opName
		case "arguments":
			if err := json.Unmarshal([]byte(*v), &arguments); err != nil {
				return err
			}
		}
	}

	var op LDAPOperation

	if arguments != nil {
		switch l.operationName {
		case "search":
			op = new(apldap.SearchOperation)
		case "":
		default:
			return fmt.Errorf("Unsupported LDAP operation %s", l.operationName)
		}

		if op != nil {
			if err := json.Unmarshal([]byte(arguments), op); err != nil {
				return err
			}
			l.arguments = op
		}
	}

	return nil
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

func (l *LDAPRequest) updateWith(other *LDAPRequest) {
	if other.Username != "" {
		l.Username = other.Username
	}

	if other.Password != "" {
		l.Password = other.Password
	}

	if other.Host != "" {
		l.Host = other.Host
	}

	if other.Port > 0 {
		l.Port = other.Port
	}
}

// Log TODO
func (l *LDAPRequest) Log(devMode bool) string {
	return "TODO"
}

// Perform TODO
func (l *LDAPRequest) Perform() Response {
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", l.Host, l.Port))
	if err != nil {
		return NewErrorResponse(aperrors.NewWrapped("[ldap] Dialing ldap endpoint", err))
	}

	defer conn.Close()
	resp, err := l.arguments.Invoke(conn)
	if err != nil {
		return NewErrorResponse(aperrors.NewWrapped("[ldap] Executing operation", err))
	}

	return resp
}
