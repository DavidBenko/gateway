package vm

import (
	"fmt"
	"os"

	"github.com/robertkrimen/otto"
)

func (p *ProxyVM) environmentGet(call otto.FunctionCall) otto.Value {
	key := call.Argument(0).String()

	value := os.Getenv(key)
	v, err := otto.ToValue(value)
	if err != nil {
		runtimeError(fmt.Sprintf("Error converting '%v' to value", value))
	}

	return v
}
