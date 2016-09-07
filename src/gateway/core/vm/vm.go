package vm

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gateway/config"
	"gateway/logreport"
	"gateway/model"

	"github.com/jmoiron/sqlx/types"
	"github.com/robertkrimen/otto"
)

var errCodeTimeout = errors.New("JavaScript took too long to execute")

// CoreVM is an Otto VM with some helper data stored alongside it.
type CoreVM struct {
	*otto.Otto
	conf                    config.ProxyServer
	LogPrint                logreport.Logf
	LogPrefix               string
	Log                     bytes.Buffer
	ProxiedRequestsDuration time.Duration
	timeout                 int64
}

// InitCoreVM initializes a new Otto VM initialized with Gateway JavaScript libraries.
func (c *CoreVM) InitCoreVM(
	vm *otto.Otto,
	logPrint logreport.Logf,
	logPrefix string,
	conf config.ProxyServer,
	proxyEndpoint *model.ProxyEndpoint,
	libraries []*model.Library,
	timeout int64,
) error {

	c.Otto = vm
	c.conf = conf
	c.LogPrint = logPrint
	c.LogPrefix = logPrefix
	c.timeout = timeout

	var scripts = make([]interface{}, 0)

	for _, library := range libraries {
		libraryCode, err := scriptFromJSONScript(library.Data)
		if err != nil {
			return err
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

	c.Set("log", c.log)

	if _, err := c.RunAll(scripts); err != nil {
		return err
	}

	return nil
}

// Run runs the given script, preventing infinite loops and very slow JS
func (c *CoreVM) Run(script interface{}) (value otto.Value, err error) {
	codeTimeout := int64(0)
	if c.timeout < 1 || c.timeout > c.conf.CodeTimeout {
		codeTimeout = c.conf.CodeTimeout
	} else {
		codeTimeout = c.timeout
	}
	defer func() {
		if caught := recover(); caught != nil {
			if caught == errCodeTimeout {
				err = errCodeTimeout
				return
			}
			panic(caught)
		}
	}()

	if c.Otto.Interrupt == nil {
		timeoutChannel := make(chan func(), 1)
		c.Otto.Interrupt = timeoutChannel

		go func() {
			time.Sleep(time.Duration(codeTimeout) * time.Second)
			timeoutChannel <- func() { panic(errCodeTimeout) }
		}()
	}

	value, err = c.Otto.Run(script)
	if err != nil {
		return value, &jsError{err, script, c.conf.NumErrorLines}
	}
	return
}

// RunAll runs all the given scripts
func (c *CoreVM) RunAll(scripts []interface{}) (value otto.Value, err error) {
	for _, script := range scripts {
		value, err = c.Run(script)
		if err != nil {
			return
		}
	}
	return
}

func (c *CoreVM) log(call otto.FunctionCall) otto.Value {
	line := call.Argument(0).String()
	c.LogPrint("%s [user] %v", c.LogPrefix, line)
	c.Log.WriteString(line + "\n")
	return otto.Value{}
}

func (c *CoreVM) runStoredJSONScript(jsonScript types.JsonText) error {
	script, err := scriptFromJSONScript(jsonScript)
	if err != nil || script == "" {
		return err
	}
	_, err = c.Run(script)
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
