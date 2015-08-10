package mongo

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"

  "gopkg.in/mgo.v2"

	"gateway/db"
)

type Conn map[string]interface{}

type Spec struct {
	mongoConn Conn
	limit int
}

func Config(confs ...db.Configurator) (db.Specifier, error) {
	spec := &Spec{}
	for _, conf := range confs {
		err := conf(spec)
		if err != nil {
			return nil, err
		}
	}
	return spec, nil
}

func Connection(s Conn) db.Configurator {
	return func(spec db.Specifier) error {
		// https://github.com/denisenkom/go-mssqldb#connection-parameters
		for _, k := range []string{
			"hosts",
			"username",
			"password",
			"database",
		} {
			if _, ok := s[k]; !ok {
				return fmt.Errorf("Mongo Config missing %q key", k)
			}
		}

		hasHost := false
		if hosts := s["hosts"]; hosts != nil {
			for _, h := range hosts.([]interface{}) {
				host := h.(map[string] interface{})
				if host["host"].(string) != "" && host["port"] != nil {
					hasHost = true
				}
			}
		}
		if !hasHost {
			return fmt.Errorf("At least one host must be defined.")
		}

		switch mongo := spec.(type) {
		case *Spec:
			mongo.mongoConn = s
			return nil
		default:
			return fmt.Errorf("Mongo Server Connection requires Conn, got %T", spec)
		}
	}
}

func PoolLimit(limit int) db.Configurator {
	return func(spec db.Specifier) error {
		switch mongo := spec.(type) {
		case *Spec:
			mongo.limit = limit
			return nil
		default:
			return fmt.Errorf("Mongo PoolLimit requires mongo.Conn, got %T", spec)
		}
	}
}

func (s *Spec) ConnectionString() string {
	conn := s.mongoConn

  //http://godoc.org/gopkg.in/mgo.v2#Dial
	var buffer bytes.Buffer
	buffer.WriteString("mongodb://")
  if conn["username"] != nil && conn["password"] != nil {
    buffer.WriteString(conn["username"].(string))
    buffer.WriteString(":")
    buffer.WriteString(conn["password"].(string))
    buffer.WriteString("@")
  }
	comma := ""
	for _, h := range conn["hosts"].([]interface{}) {
		host := h.(map[string] interface{})
    buffer.WriteString(fmt.Sprintf("%v%v:%v", comma, host["host"], host["port"]))
		comma = ","
	}
  if conn["database"] != nil {
    buffer.WriteString("/")
    buffer.WriteString(conn["database"].(string))
  }

	return buffer.String()
}

func (s *Spec) UniqueServer() string {
	conn := s.mongoConn

	keys := []string{}
	for key := range conn {
			keys = append(keys, key)
	}

	sort.Strings(keys)

	var buffer bytes.Buffer
	for _, key := range keys {
		if key == "hosts" {
			buffer.WriteString(fmt.Sprintf("%s=", key))
			for _, h := range conn[key].([]interface{}) {
				host := h.(map[string] interface{})
				buffer.WriteString(fmt.Sprintf("%v:%v,", host["host"], host["port"]))
			}
		} else {
			buffer.WriteString(fmt.Sprintf("%s=%v;", key, conn[key]))
		}
	}

	return buffer.String()
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
	mongo, err :=  mgo.Dial(s.ConnectionString())
	if err != nil {
		return nil, err
	}

	mongo.SetPoolLimit(s.limit)

	return &DB{mongo, s}, nil
}
