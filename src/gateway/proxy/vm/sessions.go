package vm

import (
	"encoding/json"
	"math"

	"github.com/robertkrimen/otto"
)

type sessionOptions struct {
	Path   string `json:"path"`
	Domain string `json:"domain"`
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int  `json:"maxAge"`
	Secure   bool `json:"secure"`
	HTTPOnly bool `json:"httpOnly"`
}

func (p *ProxyVM) sessionGet(call otto.FunctionCall) otto.Value {
	sessionName := call.Argument(0).String()
	key := call.Argument(1).String()
	session, err := p.sessionStore.Get(p.r, sessionName)
	if err != nil {
		runtimeError(err.Error())
	}
	v, err := otto.ToValue(session.Values[key])
	if err != nil {
		runtimeError(err.Error())
	}
	return v
}

func (p *ProxyVM) sessionSet(call otto.FunctionCall) otto.Value {
	sessionName := call.Argument(0).String()
	key := call.Argument(1).String()
	value := typedSessionValue(call.Argument(2))

	session, err := p.sessionStore.Get(p.r, sessionName)
	if err != nil {
		runtimeError(err.Error())
	}
	session.Values[key] = value
	session.Save(p.r, p.w)
	return otto.Value{}
}

func (p *ProxyVM) sessionIsSet(call otto.FunctionCall) otto.Value {
	sessionName := call.Argument(0).String()
	key := call.Argument(1).String()
	session, err := p.sessionStore.Get(p.r, sessionName)
	if err != nil {
		runtimeError(err.Error())
	}
	_, ok := session.Values[key]
	v, err := otto.ToValue(ok)
	if err != nil {
		runtimeError(err.Error())
	}
	return v
}

func (p *ProxyVM) sessionDelete(call otto.FunctionCall) otto.Value {
	sessionName := call.Argument(0).String()
	key := call.Argument(1).String()
	session, err := p.sessionStore.Get(p.r, sessionName)
	if err != nil {
		runtimeError(err.Error())
	}
	delete(session.Values, key)
	session.Save(p.r, p.w)
	return otto.Value{}
}

func (p *ProxyVM) sessionSetOptions(call otto.FunctionCall) otto.Value {
	sessionName := call.Argument(0).String()
	optionsString := call.Argument(1).String()

	session, err := p.sessionStore.Get(p.r, sessionName)
	if err != nil {
		runtimeError(err.Error())
	}

	var options sessionOptions
	if err := json.Unmarshal([]byte(optionsString), &options); err != nil {
		runtimeError(err.Error())
	}

	session.Options.Path = options.Path
	session.Options.Domain = options.Domain
	session.Options.MaxAge = options.MaxAge
	session.Options.Secure = options.Secure
	session.Options.HttpOnly = options.HTTPOnly

	session.Save(p.r, p.w)
	return otto.Value{}
}

func typedSessionValue(arg otto.Value) interface{} {
	if arg.IsBoolean() {
		argBool, err := arg.ToBoolean()
		if err != nil {
			runtimeError(err.Error())
		}
		return argBool
	}

	if arg.IsNumber() {
		argFloat, err := arg.ToFloat()
		if err != nil {
			runtimeError(err.Error())
		}

		argInt, err := arg.ToInteger()
		if err != nil {
			runtimeError(err.Error())
		}

		equal := func(x, y float64) bool {
			return math.Abs(x-y) <= math.SmallestNonzeroFloat64
		}

		if equal(math.Floor(argFloat), float64(argInt)) &&
			equal(math.Ceil(argFloat), float64(argInt)) {
			return argInt
		}

		return argFloat
	}

	return arg.String()
}
