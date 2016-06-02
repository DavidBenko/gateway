package redis

import (
	"errors"
	"fmt"
	"gateway/db"
)

type redisSpec interface {
	db.Specifier
}

// Spec implements db.Specifier for Redis connection parameters.
type Spec struct {
	redisSpec
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
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
		r.Database,
	)
}

func (r *Spec) UniqueServer() string {
	return r.ConnectionString()
}

func (r *Spec) NeedsUpdate(spec db.Specifier) bool {
	return false
}

func (r *Spec) NewDB() (db.DB, error) {
	return nil, nil
}

func Config(confs ...db.Configurator) (db.Specifier, error) {
	var spec redisSpec
	var ok bool

	for _, conf := range confs {
		s, err := conf(spec)
		if err != nil {
			return nil, err
		}
		spec, ok = s.(redisSpec)
		if !ok {
			return nil, fmt.Errorf("redis.Config requires Redis Specifier, got %T", s)
		}

	}
	return spec, nil
}

func Connection(c redisSpec) db.Configurator {
	return func(s db.Specifier) (db.Specifier, error) {
		if c == nil {
			return nil, errors.New("can't validate nil specifier")
		}
		message := ""

		spec, ok := c.(*Spec)
		if !ok {
			return nil, fmt.Errorf("invalid type %T", c)
		}

		if spec.Username == "" {
			message += "requires Username"
		}

		if spec.Password == "" {
			message += "; requires Password"
		}

		if spec.Host == "" {
			message += "; requires Host"
		}

		if spec.Port == 0 {
			message += "; requires Port"
		}

		if message != "" {
			message = "redis config errors: " + message
			return nil, errors.New(message)
		}

		return spec, nil
	}
}
