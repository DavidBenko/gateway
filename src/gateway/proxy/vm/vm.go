package vm

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gateway/config"
	"gateway/model"
	"gateway/sql"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx/types"

	"github.com/robertkrimen/otto"

	// Add underscore.js functionality to our VMs
	_ "github.com/robertkrimen/otto/underscore"
)

// ProxyVM is an Otto VM with some helper data stored alongside it.
type ProxyVM struct {
	*otto.Otto
	conf                    config.ProxyServer
	LogPrefix               string
	ProxiedRequestsDuration time.Duration

	w            http.ResponseWriter
	r            *http.Request
	sessionStore *sessions.CookieStore
}

// NewVM returns a new Otto VM initialized with Gateway JavaScript libraries.
func NewVM(
	logPrefix string,
	w http.ResponseWriter,
	r *http.Request,
	conf config.ProxyServer,
	db *sql.DB,
	proxyEndpoint *model.ProxyEndpoint,
) (*ProxyVM, error) {

	vm := &ProxyVM{
		otto.New(),
		conf, logPrefix, 0,
		w, r,
		nil,
	}

	var files = []string{
		"gateway.js",
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

	libraries, err := model.AllLibrariesForProxy(db, proxyEndpoint.APIID)

	if err != nil {
		return nil, err
	}
	for _, library := range libraries {
		libraryCode, err := scriptFromJSONScript(library.Data)
		if err != nil {
			return nil, err
		}
		if libraryCode != "" {
			scripts = append(scripts, libraryCode)
		}
	}

	injectEnvironment := fmt.Sprintf("var env = %s;", string(proxyEndpoint.Environment.Data))
	scripts = append(scripts, injectEnvironment)
	if conf.EnableOSEnv {
		scripts = append(scripts, osEnvironmentScript())
	}

	if err = vm.setupSessionStore(proxyEndpoint.Environment); err != nil {
		return nil, err
	}

	vm.Set("log", vm.log)
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
	log.Printf("%s [user] %v", p.LogPrefix, call.Argument(0).String())
	return otto.Value{}
}

// Panics with otto.Value are caught as runtime errors.
func runtimeError(err string) {
	errValue, _ := otto.ToValue(err)
	panic(errValue)
}

func (p *ProxyVM) runStoredJSONScript(jsonScript types.JsonText) error {
	script, err := scriptFromJSONScript(jsonScript)
	if err != nil || script == "" {
		return err
	}
	_, err = p.Run(script)
	return err
}

func scriptFromJSONScript(jsonScript types.JsonText) (string, error) {
	return strconv.Unquote(string(jsonScript))
}

func osEnvironmentScript() string {
	var keypairs []string
	for _, envPair := range os.Environ() {
		kv := strings.Split(envPair, "=")
		keypairs = append(keypairs, fmt.Sprintf("%s:%s",
			strconv.Quote(kv[0]), strconv.Quote(kv[1])))
	}

	script := fmt.Sprintf("env = _.extend({%s}, env);",
		strings.Join(keypairs, ",\n"))
	return script
}
