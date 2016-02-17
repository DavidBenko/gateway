package request

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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

	UseTLS    bool
	TLSConfig TLS

	operationName string
	arguments     apldap.Operation
	options       map[string]interface{}

	connection *apldap.ConnectionAdapter
}

// TLS configuration for a request
type TLS struct {
	PrivateKeyPassword string `json:"private_key_password"`
	ServerName         string `json:"server_name"`
	PrivateKey         string `json:"private_key"`
	Certificate        string `json:"certificate"`
}

// UnmarshalJSON is a custom method to unmarshal LDAPRequest.  A custom method
// is needed since 'arguments' is an ldap.Operation, which is not a concrete
// type, but an interface
func (l *LDAPRequest) UnmarshalJSON(data []byte) error {

	var helper struct {
		Host          string                 `json:"host"`
		Port          int                    `json:"port"`
		Username      string                 `json:"username"`
		Password      string                 `json:"password"`
		UseTLS        bool                   `json:"use_tls"`
		TLSConfig     TLS                    `json:"tls"`
		OperationName string                 `json:"operationName"`
		Options       map[string]interface{} `json:"options"`
		Arguments     *json.RawMessage       `json:"arguments"`
	}

	if err := json.Unmarshal(data, &helper); err != nil {
		return err
	}

	l.Host = helper.Host
	l.Port = helper.Port
	l.Username = helper.Username
	l.Password = helper.Password
	l.UseTLS = helper.UseTLS
	l.TLSConfig = helper.TLSConfig
	l.operationName = helper.OperationName
	l.options = helper.Options

	var op apldap.Operation

	if helper.Arguments != nil {
		switch l.operationName {
		case "search":
			op = apldap.NewSearchOperation(l.options)
		case "bind":
			op = new(apldap.BindOperation)
		case "add":
			op = new(apldap.AddOperation)
		case "delete":
			op = new(apldap.DeleteOperation)
		case "modify":
			op = new(apldap.ModifyOperation)
		case "compare":
			op = new(apldap.CompareOperation)
		case "":
		default:
			return fmt.Errorf("Unsupported LDAP operation %s", l.operationName)
		}

		if op != nil {
			if err := json.Unmarshal([]byte(*helper.Arguments), op); err != nil {
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

	if other.UseTLS && !l.UseTLS {
		l.UseTLS = true
	}

	if other.TLSConfig.Certificate != "" {
		l.TLSConfig.Certificate = other.TLSConfig.Certificate
	}

	if other.TLSConfig.PrivateKey != "" {
		l.TLSConfig.PrivateKey = other.TLSConfig.PrivateKey
	}

	if len(other.TLSConfig.PrivateKeyPassword) > 0 {
		l.TLSConfig.PrivateKeyPassword = other.TLSConfig.PrivateKeyPassword
	}

	if other.TLSConfig.ServerName != "" {
		l.TLSConfig.ServerName = other.TLSConfig.ServerName
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
	} else {
		var ldapConn *apldap.ConnectionAdapter
		var ok bool
		if ldapConn, ok = conn.(*apldap.ConnectionAdapter); ok {
			l.connection = ldapConn
		} else {
			return nil, fmt.Errorf("Expected conn to be of type *ldap.ConnectionAdapter")
		}
	}

	if l.UseTLS {
		conf, err := l.getTLSConf()
		if err != nil {
			return nil, aperrors.NewWrapped("[ldap] Getting TLS Configuration", err)
		}
		err = l.connection.Conn.StartTLS(conf)
		if err != nil {
			return nil, aperrors.NewWrapped("[ldap] Starting TLS", err)
		}
	}

	return l.connection, nil
}

func (l *LDAPRequest) getTLSConf() (*tls.Config, error) {
	conf := &tls.Config{InsecureSkipVerify: true}
	if l.TLSConfig.ServerName != "" {
		conf.ServerName = l.TLSConfig.ServerName
		conf.InsecureSkipVerify = false
	}
	if l.TLSConfig.Certificate != "" && l.TLSConfig.PrivateKey != "" {
		conf.InsecureSkipVerify = false

		var pkBytes []byte
		if len(l.TLSConfig.PrivateKeyPassword) > 0 {
			pkBlock, _ := pem.Decode([]byte(l.TLSConfig.PrivateKey))
			if pkBlock == nil {
				return nil, fmt.Errorf("No PEM data found in private key")
			}

			decryptedBytes, err := x509.DecryptPEMBlock(pkBlock, []byte(l.TLSConfig.PrivateKeyPassword))
			if err != nil {
				return nil, aperrors.NewWrapped("[ldap] Decrypting private key", err)
			}

			encodedBytes := pem.EncodeToMemory(&pem.Block{Type: pkBlock.Type, Bytes: decryptedBytes})

			pkBytes = encodedBytes
		} else {
			pkBytes = []byte(l.TLSConfig.PrivateKey)
		}

		cert, err := tls.X509KeyPair([]byte(l.TLSConfig.Certificate), pkBytes)
		if err != nil {
			return nil, aperrors.NewWrapped("[ldap] Creating TLS keypair", err)
		}

		conf.Certificates = []tls.Certificate{cert}
	}
	return conf, nil
}
