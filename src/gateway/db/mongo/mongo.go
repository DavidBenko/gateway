package mongo

import (
	"bytes"
	"fmt"
	"net/url"
	"runtime"

	"gopkg.in/mgo.v2"

	"gateway/db"
)

type Conn map[string]interface{}

type Spec struct {
	mongoConn Conn
	limit     int
}

func Config(confs ...db.Configurator) (db.Specifier, error) {
	spec := &Spec{}
	var ok bool
	for _, conf := range confs {
		s, err := conf(spec)
		if err != nil {
			return nil, err
		}
		spec, ok = s.(*Spec)
		if !ok {
			return nil, fmt.Errorf("Mongo Config expected *Spec, got %T", s)
		}
	}
	return spec, nil
}

func Connection(s Conn) db.Configurator {
	return func(spec db.Specifier) (db.Specifier, error) {
		// https://godoc.org/gopkg.in/mgo.v2
		for _, k := range []string{
			"hosts",
			"username",
			"password",
			"database",
		} {
			if _, ok := s[k]; !ok {
				return nil, fmt.Errorf("Mongo Config missing %q key", k)
			}
		}

		hasValidHost := false
		if hosts, valid := s["hosts"].([]interface{}); valid {
			for _, host := range hosts {
				if host, valid := host.(map[string]interface{}); valid {
					_host, hasHost := host["host"].(string)
					hasHost = hasHost && _host != ""
					_, hasPort := host["port"].(float64)
					if hasHost && hasPort {
						hasValidHost = true
					} else if !hasHost {
						return nil, fmt.Errorf("Host name is required")
					} else {
						return nil, fmt.Errorf("Port is required")
					}
				}
			}
		}
		if !hasValidHost {
			return nil, fmt.Errorf("At least one host must be defined.")
		}

		switch mongo := spec.(type) {
		case *Spec:
			mongo.mongoConn = s
			return mongo, nil
		default:
			return nil, fmt.Errorf("Mongo Server Connection requires Conn, got %T", spec)
		}
	}
}

func PoolLimit(limit int) db.Configurator {
	return func(spec db.Specifier) (db.Specifier, error) {
		switch mongo := spec.(type) {
		case *Spec:
			mongo.limit = limit
			return mongo, nil
		default:
			return nil, fmt.Errorf("Mongo PoolLimit requires mongo.Conn, got %T", spec)
		}
	}
}

func (s *Spec) ConnectionString() string {
	conn := s.mongoConn

	//http://godoc.org/gopkg.in/mgo.v2#Dial
	var buffer bytes.Buffer
	buffer.WriteString("mongodb://")
	buffer.WriteString(url.QueryEscape(conn["username"].(string)))
	buffer.WriteString(":")
	buffer.WriteString(url.QueryEscape(conn["password"].(string)))
	buffer.WriteString("@")

	comma := ""
	for _, h := range conn["hosts"].([]interface{}) {
		host := h.(map[string]interface{})
		buffer.WriteString(fmt.Sprintf("%v%v:%v", comma, host["host"], host["port"]))
		comma = ","
	}

	buffer.WriteString("/")
	buffer.WriteString(conn["database"].(string))

	return buffer.String()
}

func (s *Spec) UniqueServer() string {
	return s.ConnectionString()
}

func (s *Spec) NeedsUpdate(spec db.Specifier) bool {
	mongo := spec.(*Spec)
	return s.limit != mongo.limit
}

type DB struct {
	*mgo.Session
	conf *Spec
}

func (d *DB) Spec() db.Specifier {
	return d.conf
}

func mongoCloser(d *DB) {
	d.Close()
}

func (d *DB) Copy() *DB {
	copy := &DB{d.Session.Copy(), d.conf}
	runtime.SetFinalizer(copy, mongoCloser)
	return copy
}

func (d *DB) Update(s db.Specifier) error {
	spec, ok := s.(*Spec)
	if !ok {
		return fmt.Errorf("can't update Mongo Server database with \"%T\"", spec)
	}

	d.conf.limit = spec.limit
	d.Session.SetPoolLimit(spec.limit)
	return nil
}

func (s *Spec) NewDB() (db.DB, error) {
	mongo, err := mgo.Dial(s.ConnectionString())
	if err != nil {
		return nil, err
	}

	poolLimit := 16
	if s.limit <= 0 {
		poolLimit = s.limit
	}

	mongo.SetPoolLimit(poolLimit)
	d := &DB{mongo, s}
	runtime.SetFinalizer(d, mongoCloser)
	return d, nil
}
