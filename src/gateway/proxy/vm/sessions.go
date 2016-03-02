package vm

import (
	"encoding/json"
	"math"
	"time"

	aphttp "gateway/http"
	"gateway/model"
	"gateway/sql"

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

const DEFAULT_MAX_AGE = 30 * 24 * 60 * 60

type ServerStore struct {
	Header string
	UUID   string
}

type ServerSession struct {
	ID            int64
	SessionName   string `db:"session_name"`
	SessionUUID   string `db:"session_uuid"`
	MaxAge        int64  `db:"max_age"`
	Expires       int64
	SessionValues string `db:"session_values"`

	Values map[string]interface{}
}

func (p *ProxyVM) setupSessionStore(env *model.Environment) error {
	switch env.SessionType {
	case model.SessionTypeClient:
		p.Set("__ap_session_get", p.sessionGet)
		p.Set("__ap_session_set", p.sessionSet)
		p.Set("__ap_session_is_set", p.sessionIsSet)
		p.Set("__ap_session_delete", p.sessionDelete)
		p.Set("__ap_session_set_options", p.sessionSetOptions)

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
	case model.SessionTypeServer:
		p.Set("__ap_session_get", p.serverSessionGet)
		p.Set("__ap_session_set", p.serverSessionSet)
		p.Set("__ap_session_is_set", p.serverSessionIsSet)
		p.Set("__ap_session_delete", p.serverSessionDelete)
		p.Set("__ap_session_set_options", p.serverSessionSetOptions)

		header := env.SessionHeader
		if header == "" {
			header = model.SessionHeaderDefault
		}

		p.serverStore = &ServerStore{Header: header}
		if headers := p.r.Header[p.serverStore.Header]; len(headers) > 0 {
			p.serverStore.UUID = headers[0]
			p.w.Header().Add(p.serverStore.Header, headers[0])
		}
	}
	return nil
}

func (p *ProxyVM) serverSessionGet(call otto.FunctionCall) otto.Value {
	session := p.serverSession(call)
	key := call.Argument(1).String()
	v, err := otto.ToValue(session.Values[key])
	if err != nil {
		runtimeError(err.Error())
	}
	return v
}

func (p *ProxyVM) serverSessionSet(call otto.FunctionCall) otto.Value {
	session := p.serverSession(call)
	key := call.Argument(1).String()
	value := typedSessionValue(call.Argument(2))
	session.Values[key] = value
	session.Update(p.db)
	return otto.Value{}
}

func (p *ProxyVM) serverSessionIsSet(call otto.FunctionCall) otto.Value {
	session := p.serverSession(call)
	key := call.Argument(1).String()
	_, ok := session.Values[key]
	v, err := otto.ToValue(ok)
	if err != nil {
		runtimeError(err.Error())
	}
	return v
}

func (p *ProxyVM) serverSessionDelete(call otto.FunctionCall) otto.Value {
	session := p.serverSession(call)
	key := call.Argument(1).String()
	delete(session.Values, key)
	session.Update(p.db)
	return otto.Value{}
}

func (p *ProxyVM) serverSessionSetOptions(call otto.FunctionCall) otto.Value {
	session := p.serverSession(call)
	optionsString := call.Argument(1).String()
	var options sessionOptions
	if err := json.Unmarshal([]byte(optionsString), &options); err != nil {
		runtimeError(err.Error())
	}

	if options.MaxAge < 0 {
		session.Delete(p.db)
	} else if options.MaxAge > 0 {
		session.MaxAge = int64(options.MaxAge)
		session.Update(p.db)
	}

	return otto.Value{}
}

func (p *ProxyVM) serverSession(call otto.FunctionCall) *ServerSession {
	sessionName := call.Argument(0).String()

	if sessionName == "" {
		runtimeError("Sessions must have a name configured in the environment before being used.")
	}

	session := ServerSession{}
	if p.serverStore.UUID != "" {
		err := p.db.Get(&session, p.db.SQL("sessions/find"), sessionName, p.serverStore.UUID)
		if err == nil {
			err = json.Unmarshal([]byte(session.SessionValues), &session.Values)
			if err != nil {
				runtimeError(err.Error())
			}
		}
	}

	if session.ID == 0 {
		uuid, err := aphttp.NewUUID()
		if err != nil {
			runtimeError("Couldn't make UUID for server session store.")
		}
		p.w.Header().Set(p.serverStore.Header, uuid)
		p.serverStore.UUID = uuid
		session = ServerSession{
			SessionName: sessionName,
			SessionUUID: uuid,
			MaxAge:      DEFAULT_MAX_AGE,
			Expires:     time.Now().Add(DEFAULT_MAX_AGE * time.Second).Unix(),
			Values:      make(map[string]interface{}),
		}
		session.Insert(p.db)
	}

	return &session
}

func (s *ServerSession) Insert(db *sql.DB) {
	data, err := json.Marshal(s.Values)
	if err != nil {
		runtimeError(err.Error())
	}
	err = db.DoInTransaction(func(tx *sql.Tx) error {
		var err error
		s.ID, err = tx.InsertOne(tx.SQL("sessions/insert"),
			s.SessionName, s.SessionUUID,
			s.MaxAge, s.Expires, string(data))
		return err
	})
	if err != nil {
		runtimeError(err.Error())
	}
}

func (s *ServerSession) Update(db *sql.DB) {
	data, err := json.Marshal(s.Values)
	if err != nil {
		runtimeError(err.Error())
	}
	err = db.DoInTransaction(func(tx *sql.Tx) error {
		s.Expires = time.Now().Add(time.Duration(s.MaxAge) * time.Second).Unix()
		return tx.UpdateOne(tx.SQL("sessions/update"),
			s.MaxAge, s.Expires, string(data),
			s.ID, s.SessionName, s.SessionUUID)
	})
	if err != nil {
		runtimeError(err.Error())
	}
}

func (s *ServerSession) Delete(db *sql.DB) {
	err := db.DoInTransaction(func(tx *sql.Tx) error {
		return tx.DeleteOne(tx.SQL("sessions/delete"),
			s.ID, s.SessionName, s.SessionUUID)
	})
	if err != nil {
		runtimeError(err.Error())
	}
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
