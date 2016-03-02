package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gateway/config"
	"gateway/core/request"
	"gateway/db/pools"
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
) (request.Request, error) {
	if !model.IsRemoteEndpointTypeEnabled(endpoint.Type) {
		return nil, fmt.Errorf("Remote endpoint type %s is not enabled", endpoint.Type)
	}

	switch endpoint.Type {
	case model.RemoteEndpointTypeHTTP:
		return request.NewHTTPRequest(s.HTTPClient, endpoint, data)
	case model.RemoteEndpointTypeSQLServer:
		return request.NewSQLServerRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypePostgres:
		return request.NewPostgresRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeMySQL:
		return request.NewMySQLRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeMongo:
		return request.NewMongoRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeSoap:
		return request.NewSoapRequest(endpoint, data, s.SoapConf, s.OwnDb)
	case model.RemoteEndpointTypeScript:
		return request.NewScriptRequest(endpoint, data)
	}
	return nil, fmt.Errorf("%q is not a valid endpoint type", endpoint.Type)
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
