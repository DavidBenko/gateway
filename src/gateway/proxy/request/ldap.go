package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	aperrors "gateway/errors"
	"gateway/model"
	apldap "gateway/proxy/request/ldap"

	"github.com/go-ldap/ldap"
)

// LDAPRequest encapsulates a request to an LDAP server
type LDAPRequest struct {
	Host     string
	Port     int
	Username string
	Password string

	operationName string
	arguments     apldap.Operation
	options       map[string]interface{}

	connection *apldap.ConnectionAdapter
}

// UnmarshalJSON is a custom method to unmarshal LDAPRequest.  A custom method
// is needed since 'arguments' is an ldap.Operation, which is not a concrete
// type, but an interface
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
		if v == nil {
			continue
		}

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
		case "options":
			var options map[string]interface{}
			if err := json.Unmarshal([]byte(*v), &options); err != nil {
				return err
			}
			l.options = options
		}
	}

	var op apldap.Operation

	if arguments != nil {
		switch l.operationName {
		case "search":
			op = apldap.NewSearchOperation(l.options)
		case "bind":
			op = new(apldap.BindOperation)
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

// NewLDAPRequest creates a new LDAP request
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

// Log satisfies request.Request's Log method
func (l *LDAPRequest) Log(devMode bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s ldap://%s:%s@%s:%d", l.operationName, l.Username, strings.Repeat("*", len(l.Password)), l.Host, l.Port))
	if devMode {
		buffer.WriteString("\n")
		buffer.WriteString(l.arguments.PrettyString())
	}
	return buffer.String()
}

// Perform satisfies request.Request's Perform method
func (l *LDAPRequest) Perform() Response {
	if l.Username != "" && l.Password != "" {
		if err := l.connection.Conn.Bind(l.Username, l.Password); err != nil {
			return NewErrorResponse(aperrors.NewWrapped("[ldap] Invalid credentials", err))
		}
	}

	resp, err := l.arguments.Invoke(l.connection.Conn)
	if err != nil {
		return NewErrorResponse(aperrors.NewWrapped("[ldap] Executing operation", err))
	}

	return resp
}

// CreateOrReuse satisfies Initialize method on request.ReusableConnection
func (l *LDAPRequest) CreateOrReuse(conn io.Closer) (io.Closer, error) {
	if conn == nil {
		newConn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", l.Host, l.Port))
		if err != nil {
			return nil, aperrors.NewWrapped("[ldap] Dialing ldap endpoint", err)
		}

		l.connection = &apldap.ConnectionAdapter{newConn}

		return l.connection, nil
	}

	if ldapConn, ok := conn.(*apldap.ConnectionAdapter); ok {
		l.connection = ldapConn
		return ldapConn, nil
	}

	return nil, fmt.Errorf("Expected conn to be of type *ldap.ConnectionAdapter")
}
