package core

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gateway/config"
	"gateway/core/request"
	"gateway/core/vm"
	"gateway/core/vm/advanced"
	"gateway/core/vm/conversion"
	"gateway/core/vm/crypto"
	"gateway/core/vm/encoding"
	"gateway/db/pools"
	aperrors "gateway/errors"
	"gateway/logreport"
	"gateway/model"
	"gateway/push"
	"gateway/smtp"
	sql "gateway/sql"
	statssql "gateway/stats/sql"
	"gateway/store"

	"github.com/robertkrimen/otto"

	// Add underscore.js functionality to our VMs
	_ "github.com/robertkrimen/otto/underscore"
)

const (
	HttpRequest      = "http"
	RedisRequest     = "redis"
	SqlServerRequest = "sqlserver"
	PostgresRequest  = "postgres"
	MySqlRequest     = "mysql"
	MongoRequest     = "mongo"
	SoapRequest      = "soap"
	LdapRequest      = "ldap"
	HanaRequest      = "hana"
	StoreRequest     = "store"
	PushRequest      = "push"
	SmtpRequest      = "smtp"
	JobRequest       = "job"
	KeyRequest       = "key"
	ScriptRequest    = "script"
	DockerRequest    = "docker"
)

type Core struct {
	DevMode               bool
	HTTPClient            *http.Client
	InsecureHTTPClient    *http.Client
	DBPools               *pools.Pools
	OwnDb                 *sql.DB // in-application datastore
	SoapConf              config.Soap
	DockerConf            config.Docker
	Store                 store.Store
	Push                  *push.PushPool
	Smtp                  *smtp.SmtpPool
	VMKeyStore            vm.DataSource
	VMRemoteEndpointStore vm.DataSource
	StatsDb               *statssql.SQL
	Conf                  config.Configuration
}

func NewCore(conf config.Configuration, ownDb *sql.DB, statsDb *statssql.SQL) *Core {
	httpTimeout := time.Duration(conf.Proxy.HTTPTimeout) * time.Second

	keyStore := vm.NewKeyStore(ownDb, int(conf.Proxy.KeyCacheSize))
	ownDb.RegisterListener(keyStore)

	endpointStore := vm.NewRemoteEndpointStore(ownDb, int(conf.Proxy.RemoteEndpointCacheSize))
	ownDb.RegisterListener(endpointStore)

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
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Core{
		DevMode:               conf.DevMode(),
		HTTPClient:            &http.Client{Timeout: httpTimeout},
		InsecureHTTPClient:    &http.Client{Timeout: httpTimeout, Transport: tr},
		DBPools:               pools,
		OwnDb:                 ownDb,
		SoapConf:              conf.Soap,
		DockerConf:            conf.Docker,
		Store:                 objectStore,
		Push:                  push.NewPushPool(conf.Push),
		Smtp:                  smtp.NewSmtpPool(),
		VMKeyStore:            keyStore,
		VMRemoteEndpointStore: endpointStore,
		StatsDb:               statsDb,
		Conf:                  conf,
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

	generic := &struct {
		Type string `json:"__type"`
	}{}

	if err := json.Unmarshal(*data, generic); err != nil {
		return nil, fmt.Errorf("unable to determine request type: %v", err)
	}

	invalidTypeErrorMessage := func(expected string, got string) error {
		return fmt.Errorf("mismatched request types: expected %s got %s", expected, got)
	}

	switch endpoint.Type {
	case model.RemoteEndpointTypeHTTP:
		if generic.Type != HttpRequest {
			return nil, invalidTypeErrorMessage(HttpRequest, generic.Type)
		}
		return request.NewHTTPRequest(s.HTTPClient, s.InsecureHTTPClient, endpoint, data)
	case model.RemoteEndpointTypeSQLServer:
		if generic.Type != SqlServerRequest {
			return nil, invalidTypeErrorMessage(SqlServerRequest, generic.Type)
		}
		return request.NewSQLServerRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypePostgres:
		if generic.Type != PostgresRequest {
			return nil, invalidTypeErrorMessage(PostgresRequest, generic.Type)
		}
		return request.NewPostgresRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeMySQL:
		if generic.Type != MySqlRequest {
			return nil, invalidTypeErrorMessage(MySqlRequest, generic.Type)
		}
		return request.NewMySQLRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeMongo:
		if generic.Type != MongoRequest {
			return nil, invalidTypeErrorMessage(MongoRequest, generic.Type)
		}
		return request.NewMongoRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeSoap:
		if generic.Type != SoapRequest {
			return nil, invalidTypeErrorMessage(SoapRequest, generic.Type)
		}
		return request.NewSoapRequest(endpoint, data, s.SoapConf, s.OwnDb)
	case model.RemoteEndpointTypeScript:
		if generic.Type != ScriptRequest {
			return nil, invalidTypeErrorMessage(ScriptRequest, generic.Type)
		}
		return request.NewScriptRequest(endpoint, data)
	case model.RemoteEndpointTypeStore:
		if generic.Type != StoreRequest {
			return nil, invalidTypeErrorMessage(StoreRequest, generic.Type)
		}
		return request.NewStoreRequest(s.Store, endpoint, data)
	case model.RemoteEndpointTypeLDAP:
		if generic.Type != LdapRequest {
			return nil, invalidTypeErrorMessage(LdapRequest, generic.Type)
		}
		r, e := request.NewLDAPRequest(endpoint, data)
		// cache connections in the connections map for later use within the same proxy endpoint workflow
		conn, err := r.CreateOrReuse(connections[endpoint.ID])
		if err != nil {
			return nil, aperrors.NewWrapped("[requests.go] initializing sticky connection", err)
		}
		connections[endpoint.ID] = conn
		return r, e
	case model.RemoteEndpointTypePush:
		if generic.Type != PushRequest {
			return nil, invalidTypeErrorMessage(PushRequest, generic.Type)
		}
		return request.NewPushRequest(endpoint, data, s.Push, s.OwnDb)
	case model.RemoteEndpointTypeHana:
		if generic.Type != HanaRequest {
			return nil, invalidTypeErrorMessage(HanaRequest, generic.Type)
		}
		return request.NewHanaRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeRedis:
		if generic.Type != RedisRequest {
			return nil, invalidTypeErrorMessage(RedisRequest, generic.Type)
		}
		return request.NewRedisRequest(s.DBPools, endpoint, data)
	case model.RemoteEndpointTypeSMTP:
		if generic.Type != SmtpRequest {
			return nil, invalidTypeErrorMessage(SmtpRequest, generic.Type)
		}
		return request.NewSmtpRequest(s.Smtp, endpoint, data)
	case model.RemoteEndpointTypeDocker:
		if generic.Type != DockerRequest {
			return nil, invalidTypeErrorMessage(DockerRequest, generic.Type)
		}
		return request.NewDockerRequest(endpoint, data, s.DockerConf)
	case model.RemoteEndpointTypeJob:
		if generic.Type != JobRequest {
			return nil, invalidTypeErrorMessage(JobRequest, generic.Type)
		}
		return request.NewJobRequest(s.OwnDb, endpoint, s.ExecuteJob, data)
	case model.RemoteEndpointTypeKey:
		if generic.Type != KeyRequest {
			return nil, invalidTypeErrorMessage(KeyRequest, generic.Type)
		}
		return request.NewKeyRequest(s.OwnDb, endpoint, data)
	default:
		return nil, fmt.Errorf("%q is not a valid endpoint type", endpoint.Type)
	}
}

func VMCopy(accountID int64, keySource vm.DataSource, endpointSource vm.DataSource, prepare advanced.RequestPreparer) *otto.Otto {
	vm := shared.Copy()
	crypto.IncludeSigning(vm, accountID, keySource)
	crypto.IncludeEncryption(vm, accountID, keySource)
	advanced.IncludePerform(vm, accountID, endpointSource, prepare)
	return vm
}

var shared = func() *otto.Otto {
	vm := otto.New()

	var files = []string{
		"gateway.js",
		"sessions.js",
		"crypto.js",
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

	conversion.IncludeConversion(vm)
	conversion.IncludePath(vm)
	crypto.IncludeHashing(vm)
	crypto.IncludeAes(vm)
	crypto.IncludeRand(vm)
	encoding.IncludeEncoding(vm)

	return vm
}()
