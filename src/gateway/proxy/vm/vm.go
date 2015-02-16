package vm

import (
	"log"
	"net/http"
	"time"

	"gateway/config"

	"github.com/gorilla/sessions"

	"github.com/robertkrimen/otto"

	// Add underscore.js functionality to our VMs
	_ "github.com/robertkrimen/otto/underscore"
)

// ProxyVM is an Otto VM with some helper data stored alongside it.
type ProxyVM struct {
	*otto.Otto
	conf                    config.ProxyServer
	RequestID               string
	ProxiedRequestsDuration time.Duration

	/* TODO: Do both of the following get removed? */
	scripts           map[string]*otto.Script
	includedLibraries []string

	w            http.ResponseWriter
	r            *http.Request
	sessionStore *sessions.CookieStore
}

// NewVM returns a new Otto VM initialized with Gateway JavaScript libraries.
func NewVM(
	requestID string,
	w http.ResponseWriter,
	r *http.Request,
	conf config.ProxyServer,
	proxyScripts map[string]*otto.Script,
) (*ProxyVM, error) {

	var files = []string{
		"gateway.js",
		"environments.js",
		"sessions.js",
		"call.js",
		"http/request.js",
		"http/response.js",
	}
	var scripts = make([]interface{}, 0)
	for _, filename := range files {
		fileJS, err := Asset(filename)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, fileJS)
	}

	vm := &ProxyVM{
		otto.New(),
		conf, requestID, 0,
		proxyScripts, []string{},
		w, r,
		nil,
	}

	/* FIXME: Need to move keys to Environment for multi-tenant, not config */
	if conf.AuthKey != "" {
		sessionConfig := [][]byte{[]byte(conf.AuthKey)}
		if conf.EncryptionKey != "" {
			sessionConfig = append(sessionConfig, []byte(conf.EncryptionKey))
		}
		vm.sessionStore = sessions.NewCookieStore(sessionConfig...)
	}

	/* TODO: Bind to objects? & evaluate usage */
	vm.Set("__ap_log", vm.log)                        /* log("foo") instead? */
	vm.Set("__ap_environment_get", vm.environmentGet) /* env("key") instead? */
	vm.Set("__ap_session_get", vm.sessionGet)
	vm.Set("__ap_session_set", vm.sessionSet)
	vm.Set("__ap_session_is_set", vm.sessionIsSet)
	vm.Set("__ap_session_delete", vm.sessionDelete)
	vm.Set("__ap_session_set_options", vm.sessionSetOptions)

	if _, err := vm.RunAll(scripts); err != nil {
		return nil, err
	}

	return vm, nil
}

// RunAll runs all the given scripts
func (p *ProxyVM) RunAll(scripts []interface{}) (value otto.Value, err error) {
	for _, script := range scripts {
		value, err = p.Run(script)
		if err != nil {
			return
		}
	}
	return
}

func (p *ProxyVM) log(call otto.FunctionCall) otto.Value {
	log.Printf("%s [req %s] [user] %v", config.Proxy, p.RequestID, call.Argument(0).String())
	return otto.Value{}
}

// Panics with otto.Value are caught as runtime errors.
func runtimeError(err string) {
	errValue, _ := otto.ToValue(err)
	panic(errValue)
}
