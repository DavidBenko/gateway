package vm

import (
	"fmt"

	"gateway/model"

	"github.com/robertkrimen/otto"
)

func (p *ProxyVM) environmentGet(call otto.FunctionCall) otto.Value {
	key := call.Argument(0).String()

	value, ok := p.environmentValue(p.conf.Environment, key)
	if value == nil && !ok {
		value, ok = p.environmentValue(p.conf.EnvironmentDefault, key)
	}
	if !ok {
		runtimeError(fmt.Sprintf("There is no environment value named '%v'", key))
	}

	v, err := otto.ToValue(value)
	if err != nil {
		runtimeError(fmt.Sprintf("Error converting '%v' to value", value))
	}

	return v
}

func (p *ProxyVM) environmentValue(envName, key string) (interface{}, bool) {
	if envName != "" {
		env, err := p.db.Find(&model.Environment{}, "Name", envName)
		if err != nil {
			runtimeError(fmt.Sprintf("There is no environment named '%s'", envName))
		}
		value, ok := env.(*model.Environment).Values[key]
		return value, ok
	}
	return nil, false
}
