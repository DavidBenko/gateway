package vm

import (
	"encoding/json"
	"math"

	"gateway/model"

	"github.com/gorilla/sessions"
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

func (p *ProxyVM) setupSessionStore(env *model.Environment) error {
	if env.SessionAuthKey == "" {
		return nil
	}

	rotating := (env.SessionAuthKeyRotate != "")

	sessionConfig := [][]byte{[]byte(env.SessionAuthKey)}
	if env.SessionEncryptionKey != "" {
		sessionConfig = append(sessionConfig, []byte(env.SessionEncryptionKey))
	} else if rotating {
		sessionConfig = append(sessionConfig, nil)
	}
	if rotating {
		sessionConfig = append(sessionConfig, []byte(env.SessionAuthKeyRotate))
		if env.SessionEncryptionKeyRotate != "" {
			sessionConfig = append(sessionConfig, []byte(env.SessionEncryptionKeyRotate))
		}
	}
	p.sessionStore = sessions.NewCookieStore(sessionConfig...)
	return nil
}

func (p *ProxyVM) sessionGet(call otto.FunctionCall) otto.Value {
	session := p.session(call)
	key := call.Argument(1).String()
	v, err := otto.ToValue(session.Values[key])
	if err != nil {
		runtimeError(err.Error())
	}
	return v
}

func (p *ProxyVM) sessionSet(call otto.FunctionCall) otto.Value {
	session := p.session(call)
	key := call.Argument(1).String()
	value := typedSessionValue(call.Argument(2))
	session.Values[key] = value
	session.Save(p.r, p.w)
	return otto.Value{}
}

func (p *ProxyVM) sessionIsSet(call otto.FunctionCall) otto.Value {
	session := p.session(call)
	key := call.Argument(1).String()
	_, ok := session.Values[key]
	v, err := otto.ToValue(ok)
	if err != nil {
		runtimeError(err.Error())
	}
	return v
}

func (p *ProxyVM) sessionDelete(call otto.FunctionCall) otto.Value {
	session := p.session(call)
	key := call.Argument(1).String()
	delete(session.Values, key)
	session.Save(p.r, p.w)
	return otto.Value{}
}

func (p *ProxyVM) sessionSetOptions(call otto.FunctionCall) otto.Value {
	session := p.session(call)
	optionsString := call.Argument(1).String()
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

func (p *ProxyVM) session(call otto.FunctionCall) *sessions.Session {
	sessionName := call.Argument(0).String()

	if sessionName == "" {
		runtimeError("Sessions must have a name configured in the environment before being used.")
	}

	if p.sessionStore == nil {
		runtimeError("Sessions must have at least one auth key configured in the environment before being used.")
	}

	session, err := p.sessionStore.Get(p.r, sessionName)
	if err != nil {
		runtimeError(err.Error())
	}

	return session
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
