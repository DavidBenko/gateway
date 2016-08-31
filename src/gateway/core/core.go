package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gateway/config"
	"gateway/core/conversion"
	"gateway/core/request"
	"gateway/db/pools"
	aperrors "gateway/errors"
	"gateway/logreport"
	"gateway/model"
	"gateway/push"
	"gateway/smtp"
	sql "gateway/sql"
	"gateway/store"

	"github.com/robertkrimen/otto"

	// Add underscore.js functionality to our VMs
	_ "github.com/robertkrimen/otto/underscore"
)

type Core struct {
	HTTPClient *http.Client
	DBPools    *pools.Pools
	OwnDb      *sql.DB // in-application datastore
	SoapConf   config.Soap
	DockerConf config.Docker
	Store      store.Store
	Push       *push.PushPool
	Smtp       *smtp.SmtpPool
}

func (s *Core) PrepareRequest(
	endpoint *model.RemoteEndpoint,
	data *json.RawMessage,
	connections map[int64]io.Closer,
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
	case model.RemoteEndpointTypeStore:
		return request.NewStoreRequest(s.Store, endpoint, data)
	case model.RemoteEndpointTypeLDAP:
		r, e := request.NewLDAPRequest(endpoint, data)
		// cache connections in the connections map for later use within the same proxy endpoint workflow
		conn, err := r.CreateOrReuse(connections[endpoint.ID])
		if err != nil {
			return nil, aperrors.NewWrapped("[requests.go] initializing sticky connection", err)
		}
		connections[endpoint.ID] = conn
		return r, e
	case model.RemoteEndpointTypePush:
		return request.NewPushRequest(endpoint, data, s.Push, s.OwnDb)
	case model.RemoteEndpointTypeHana:
		return request.NewHanaRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeRedis:
		return request.NewRedisRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeSMTP:
		return request.NewSmtpRequest(s.Smtp, endpoint, data)
	case model.RemoteEndpointTypeDocker:
		return request.NewDockerRequest(endpoint, data, s.DockerConf)
	default:
		return nil, fmt.Errorf("%q is not a valid endpoint type", endpoint.Type)
	}
}

func VMCopy() *otto.Otto {
	return shared.Copy()
}

var shared = func() *otto.Otto {
	vm := otto.New()

	conversion.IncludeConversion(vm)

	var files = []string{
		"gateway.js",
		"sessions.js",
		"call.js",
		"http/request.js",
		"http/response.js",
		"conversion/json.js",
		"conversion/xml.js",
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
