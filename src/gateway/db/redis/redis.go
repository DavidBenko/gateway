package redis

import (
	"fmt"
	"gateway/db"
)

// Spec implements db.Specifier for Redis connection parameters.
type Spec struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
	limit    int
}

// ConnectionString returns the redis connection string derived from the
// redis.Spec.
func (r *Spec) ConnectionString() string {
	return fmt.Sprintf("redis://%s:%s@%s:%d/%s",
		r.Username,
		r.Password,
		r.Host,
		r.Port,
		r.Database)
}

func (r *Spec) UniqueServer() string {
	return r.ConnectionString()
}

func (r *Spec) NeedsUpdate(spec db.Specifier) bool {
	redis := spec.(*Spec)
	return r.limit != redis.limit
}

func (r *Spec) NewDB() (db.DB, error) {
	return nil, nil
}

func Config(confs ...db.Configurator) (db.Specifier, error) {
	spec := &spec{}
	ok := false

	for _, conf := range confs {
		s, err := conf(spec)
		if err != nil {
			return nil, err
		}
		spec, ok = s.(*Spec)
		if !ok {
			return nil, fmt.Errorf("Redis Config expected *Spec, got %T", s)
		}
	}

	return spec, nil
}
