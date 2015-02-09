package vm

import (
	"fmt"
	"os"

	"github.com/robertkrimen/otto"
)

/**
 * TODO: This needs to be disabled by default
 * We can optionally expose to VMs based on a config setting; i.e. clients can
 * choose to leak ENV for private installations, but it should be disabled by
 * default for multi-tenant solutions.
 */
func (p *ProxyVM) environmentGet(call otto.FunctionCall) otto.Value {
	key := call.Argument(0).String()

	value := os.Getenv(key)
	v, err := otto.ToValue(value)
	if err != nil {
		runtimeError(fmt.Sprintf("Error converting '%v' to value", value))
	}

	return v
}
