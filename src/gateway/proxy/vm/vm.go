package vm

import (
	"errors"
	"net/http"

	"gateway/config"
	"gateway/core"
	"gateway/core/vm"
	"gateway/logreport"
	"gateway/model"
	"gateway/sql"

	"github.com/gorilla/sessions"

	// Add underscore.js functionality to our VMs
	"github.com/robertkrimen/otto"
)

var errCodeTimeout = errors.New("JavaScript took too long to execute")

// ProxyVM is an Otto VM with some helper data stored alongside it.
type ProxyVM struct {
	vm.CoreVM
	w            http.ResponseWriter
	r            *http.Request
	sessionStore *sessions.CookieStore
	serverStore  *ServerStore
	db           *sql.DB
}

// NewVM returns a new Otto VM initialized with Gateway JavaScript libraries.
func NewVM(
	logPrint logreport.Logf,
	logPrefix string,
	w http.ResponseWriter,
	r *http.Request,
	conf config.ProxyServer,
	db *sql.DB,
	proxyEndpoint *model.ProxyEndpoint,
	libraries []*model.Library,
	timeout int64,
	keyStore *vm.KeyStore,
) (*ProxyVM, error) {
	vm := &ProxyVM{
		w:  w,
		r:  r,
		db: db,
	}
	vm.InitCoreVM(core.VMCopy(proxyEndpoint.AccountID, keyStore), logPrint, logPrefix, &conf, proxyEndpoint, libraries, timeout)

	if err := vm.setupSessionStore(proxyEndpoint.Environment); err != nil {
		return nil, err
	}

	return vm, nil
}

// Panics with otto.Value are caught as runtime errors.
func runtimeError(err string) {
	errValue, _ := otto.ToValue(err)
	panic(errValue)
}
