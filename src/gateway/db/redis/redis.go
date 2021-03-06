package redis

import (
	"errors"
	"fmt"
	"gateway/db"
	"gateway/logreport"

	redigo "github.com/garyburd/redigo/redis"
)

type redisSpec interface {
	db.Specifier
}

// Spec implements db.Specifier for Redis connection parameters.
type Spec struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Database  string `json:"database"`
	maxActive int
	maxIdle   int
}

// ConnectionString returns the redis connection string derived from the
// redis.Spec.
func (r *Spec) ConnectionString() string {
	s := "redis://"

	if r.Username != "" {
		s += r.Username
	}
	if r.Password != "" {
		s += fmt.Sprintf(":%s@", r.Password)
	}

	return fmt.Sprintf("%s%s:%d/%s",
		s,
		r.Host,
		r.Port,
		r.Database,
	)
}

func (r *Spec) UniqueServer() string {
	return r.ConnectionString()
}

func (r *Spec) NeedsUpdate(spec db.Specifier) bool {
	if spec == nil {
		logreport.Panicf("tried to compare to nil db.Specifier!")
	}
	if rSpec, ok := spec.(*Spec); ok {
		return rSpec.maxActive != r.maxActive || rSpec.maxIdle != r.maxIdle
	}
	logreport.Panicf("tried to compare wrong database kinds: Redis and %T", spec)
	return false
}

func (r *Spec) UpdateWith(spec *Spec) error {
	if spec == nil {
		return errors.New("cannot update Redis with a nil Specifier")
	}

	if err := spec.validate(); err != nil {
		return err
	}

	*r = *spec
	return nil
}

func (r *Spec) validate() error {
	message := ""

	if r.Host == "" {
		message += "requires Host;"
	}

	if r.Port == 0 {
		message += " requires Port"
	}

	if message != "" {
		message = "redis config errors: " + message
		return errors.New(message)
	}
	return nil
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

		spec, ok := c.(*Spec)
		if !ok {
			return nil, fmt.Errorf("invalid type %T", c)
		}

		err := spec.validate()
		if err != nil {
			return nil, err
		}

		return spec, nil
	}
}

func MaxActive(limit int) db.Configurator {
	return func(s db.Specifier) (db.Specifier, error) {
		switch redis := s.(type) {
		case *Spec:
			redis.maxActive = limit
			return redis, nil
		default:
			return nil, fmt.Errorf("Redis MaxActive requires redis.Spec, got %T", s)
		}
	}
}

func MaxIdle(idle int) db.Configurator {
	return func(s db.Specifier) (db.Specifier, error) {
		switch redis := s.(type) {
		case *Spec:
			redis.maxIdle = idle
			return redis, nil
		default:
			return nil, fmt.Errorf("Redis MaxIdle requires redis.Spec, got %T", s)
		}
	}
}

type DB struct {
	Pool *redigo.Pool
	conf *Spec
}

func (r *Spec) NewDB() (db.DB, error) {
	pool := newPool(r)
	db := &DB{pool, r}
	return db, nil
}

// newPool returns a redigo.pool struct. Since redigo does not implement
// a connection pooling mechanism under the hood, we have to create our own.
func newPool(r *Spec) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:   r.maxIdle,
		MaxActive: r.maxActive,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.DialURL(r.ConnectionString())
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}

func (d *DB) Spec() db.Specifier {
	return d.conf
}

func (d *DB) Update(s db.Specifier) error {
	spec, ok := s.(*Spec)
	if !ok {
		return fmt.Errorf("can't update Redis with %T", spec)
	}

	if err := spec.validate(); err != nil {
		return err
	}

	d.conf.maxActive = spec.maxActive
	d.conf.maxIdle = spec.maxIdle

	// update underlying pool
	d.Pool.MaxActive = spec.maxActive
	d.Pool.MaxIdle = spec.maxIdle

	return nil
}
