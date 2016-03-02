package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gateway/config"
	"gateway/core/request"
	"gateway/db/pools"
	aperrors "gateway/errors"
	"gateway/logreport"
	"gateway/model"
	sql "gateway/sql"

	"github.com/robertkrimen/otto"

	// Add underscore.js functionality to our VMs
	_ "github.com/robertkrimen/otto/underscore"
)

type Core struct {
	HTTPClient *http.Client
	DBPools    *pools.Pools
	OwnDb      *sql.DB // in-application datastore
	SoapConf   config.Soap
}

func (s *Core) PrepareRequest(
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
		r, e = request.NewHTTPRequest(s.HTTPClient, endpoint, data)
	case model.RemoteEndpointTypeSQLServer:
		r, e = request.NewSQLServerRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypePostgres:
		r, e = request.NewPostgresRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeMySQL:
		r, e = request.NewMySQLRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeMongo:
		r, e = request.NewMongoRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeSoap:
		r, e = request.NewSoapRequest(endpoint, data, s.SoapConf, s.OwnDb)
	case model.RemoteEndpointTypeScript:
		r, e = request.NewScriptRequest(endpoint, data)
	case model.RemoteEndpointTypeLDAP:
		l, lErr := request.NewLDAPRequest(endpoint, data)
		// cache connections in the connections map for later use within the same proxy endpoint workflow
		conn, err := l.CreateOrReuse(connections[endpoint.ID])
		if err != nil {
			return nil, aperrors.NewWrapped("[requests.go] initializing sticky connection", err)
		}
		connections[endpoint.ID] = conn
		r, e = l, lErr
	default:
		return nil, fmt.Errorf("%q is not a valid endpoint type", endpoint.Type)
	}

	if e != nil {
		return r, e
	}

	return r, nil
}

func VMCopy() *otto.Otto {
	return shared.Copy()
}

var shared = func() *otto.Otto {
	vm := otto.New()

	var files = []string{
		"gateway.js",
		"sessions.js",
		"call.js",
		"http/request.js",
		"http/response.js",
	}
	for _, filename := range files {
		fileJS, err := Asset(filename)
		if err != nil {
			logreport.Fatal(err)
		}

		_, err = vm.Run(fileJS)
		if err != nil {
			logreport.Fatal(err)
		}
	}

	return vm
}()
