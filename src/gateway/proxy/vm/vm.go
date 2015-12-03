package vm

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gateway/config"
	"gateway/logger"
	"gateway/model"
	"gateway/sql"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx/types"

	"github.com/robertkrimen/otto"

	// Add underscore.js functionality to our VMs
	_ "github.com/robertkrimen/otto/underscore"
)

var errCodeTimeout = errors.New("JavaScript took too long to execute")

// ProxyVM is an Otto VM with some helper data stored alongside it.
type ProxyVM struct {
	*otto.Otto
	conf                    config.ProxyServer
	Logger                  *log.Logger
	LogPrefix               string
	Log                     bytes.Buffer
	ProxiedRequestsDuration time.Duration

	w            http.ResponseWriter
	r            *http.Request
	sessionStore *sessions.CookieStore
}

// NewVM returns a new Otto VM initialized with Gateway JavaScript libraries.
func NewVM(
	logger *log.Logger,
	logPrefix string,
	w http.ResponseWriter,
	r *http.Request,
	conf config.ProxyServer,
	db *sql.DB,
	proxyEndpoint *model.ProxyEndpoint,
	libraries []*model.Library,
) (*ProxyVM, error) {

	vm := &ProxyVM{
		Otto:                    shared.Copy(),
		conf:                    conf,
		Logger:                  logger,
		LogPrefix:               logPrefix,
		ProxiedRequestsDuration: 0,
		w: w,
		r: r,
	}

	var scripts = make([]interface{}, 0)

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

	if err := vm.setupSessionStore(proxyEndpoint.Environment); err != nil {
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

// Run runs the given script, preventing infinite loops and very slow JS
func (p *ProxyVM) Run(script interface{}) (value otto.Value, err error) {
	defer func() {
		if caught := recover(); caught != nil {
			if caught == errCodeTimeout {
				err = errCodeTimeout
				return
			}
			panic(caught)
		}
	}()

	if p.Otto.Interrupt == nil {
		timeoutChannel := make(chan func(), 1)
		p.Otto.Interrupt = timeoutChannel

		go func() {
			time.Sleep(time.Duration(p.conf.CodeTimeout) * time.Second)
			timeoutChannel <- func() { panic(errCodeTimeout) }
		}()
	}

	value, err = p.Otto.Run(script)
	if err != nil {
		return value, &jsError{err, script, p.conf.NumErrorLines}
	}
	return
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
	line := call.Argument(0).String()
	p.Logger.Printf("%s [user] %v", p.LogPrefix, line)
	p.Log.WriteString(line + "\n")
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
			logger.Fatal(err)
		}

		_, err = vm.Run(fileJS)
		if err != nil {
			logger.Fatal(err)
		}
	}

	return vm
}()
