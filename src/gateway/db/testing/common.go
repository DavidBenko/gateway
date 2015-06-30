package testing

import (
	"gateway/db"
	"gateway/db/pools"
)

func Connect(p pools.ServerPool, specs []db.Specifier) ([]db.DB, error) {
	var dbs = make([]db.DB, 0)
	for _, spec := range specs {
		d, err := pools.Connect(p, spec)
		if err != nil {
			return nil, err
		}
		dbs = append(dbs, d)
	}
	return dbs, nil
}

// MakePool makes a ServerPool which can be used for testing.
func MakePool() *ServerPool {
	return &ServerPool{DBs: make(map[string]db.DB)}
}
