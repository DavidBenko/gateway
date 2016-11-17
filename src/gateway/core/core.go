package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gateway/config"
	"gateway/core/conversion"
	"gateway/core/ottocrypto"
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
	DevMode    bool
	HTTPClient *http.Client
	DBPools    *pools.Pools
	OwnDb      *sql.DB // in-application datastore
	SoapConf   config.Soap
	DockerConf config.Docker
	Store      store.Store
	Push       *push.PushPool
	Smtp       *smtp.SmtpPool
	KeyStore   *KeyStore
	Conf       config.Configuration
}

func NewCore(conf config.Configuration, ownDb *sql.DB) *Core {
	httpTimeout := time.Duration(conf.Proxy.HTTPTimeout) * time.Second

	keyStore := NewKeyStore(ownDb)
	ownDb.RegisterListener(keyStore)

	pools := pools.MakePools()
	ownDb.RegisterListener(pools)

	// Configure the object store
	objectStore, err := store.Configure(conf.Store)
	if err != nil {
		logreport.Fatalf("Unable to configure the object store: %v", err)
	}
	err = objectStore.Migrate()
	if err != nil {
		logreport.Fatalf("Unable to migrate the object store: %v", err)
	}

	return &Core{
		DevMode:    conf.DevMode(),
		HTTPClient: &http.Client{Timeout: httpTimeout},
		DBPools:    pools,
		OwnDb:      ownDb,
		SoapConf:   conf.Soap,
		DockerConf: conf.Docker,
		Store:      objectStore,
		Push:       push.NewPushPool(conf.Push),
		Smtp:       smtp.NewSmtpPool(),
		KeyStore:   keyStore,
		Conf:       conf,
	}
}

func (c *Core) Shutdown() {
	c.Store.Shutdown()
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
	case model.RemoteEndpointTypeJob:
		return request.NewJobRequest(s.OwnDb, endpoint, s.ExecuteJob, data)
	case model.RemoteEndpointTypeKey:
		return request.NewKeyRequest(endpoint, data)
	default:
		return nil, fmt.Errorf("%q is not a valid endpoint type", endpoint.Type)
	}
}

func VMCopy(accountID int64, keySource ottocrypto.KeyDataSource) *otto.Otto {
	vm := shared.Copy()
	ottocrypto.IncludeSigning(vm, accountID, keySource)
	ottocrypto.IncludeEncryption(vm, accountID, keySource)
	return vm
}

var shared = func() *otto.Otto {
	vm := otto.New()

	conversion.IncludeConversion(vm)
	conversion.IncludePath(vm)
	ottocrypto.IncludeHashing(vm)

	var files = []string{
		"gateway.js",
		"sessions.js",
		"crypto.js",
		"call.js",
		"http/request.js",
		"http/response.js",
		"conversion/json.js",
		"conversion/xml.js",
		"encoding.js",
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

	ottocrypto.IncludeAes(vm)

	return vm
}()
