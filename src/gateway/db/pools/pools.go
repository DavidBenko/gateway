package pools

import (
	"fmt"

	"gateway/db"
	"gateway/db/mongo"
	"gateway/db/redis"
	"gateway/db/sql"
)

// Pools handles concurrent access to databases with connection pools.
type Pools struct {
	// Pools must remain threadsafe!
	sqlsPool  *sqlPool
	pqPool    *sqlPool
	mySqlPool *sqlPool
	mongoPool *mongoPool
	redisPool *redisPool
}

// poolForSpec returns the correct pool for the given db.Specifier.
func (p *Pools) poolForSpec(spec db.Specifier) (ServerPool, error) {
	switch spec.(type) {
	case *sql.SQLServerSpec:
		return p.sqlsPool, nil
	case *sql.PostgresSpec:
		return p.pqPool, nil
	case *sql.MySQLSpec:
		return p.mySqlPool, nil
	case *mongo.Spec:
		return p.mongoPool, nil
	case *redis.Spec:
		return p.redisPool, nil
	default:
		return nil, fmt.Errorf("no pool defined for spec type %T", spec)
	}
}

// ServerPool manages a map of dbs and has a RWMutex to control access.
type ServerPool interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()

	Get(db.Specifier) (db.DB, bool)
	Put(db.Specifier, db.DB)
	Delete(db.Specifier)

	// Iterator should return a full, closed, buffered channel of
	// the Specifiers for each db.DB in the ServerPool.
	Iterator() <-chan db.Specifier
}

// MakePools returns a new Pools with initialized sub-pools.
func MakePools() *Pools {
	return &Pools{
		sqlsPool:  makeSqlPool(),
		pqPool:    makeSqlPool(),
		mySqlPool: makeSqlPool(),
		mongoPool: makeMongoPool(),
		redisPool: makeRedisPool(),
	}
}

// Connect returns a connection to a database with the correct type.  If
// necessary, the connection will be generated.
//
// Usage:
//
// import sql "gateway/db/sql"
//
// p.Connect(sql.Config(
//         sql.Connection(conn),
//         sql.MaxOpenIdle(100, 10),
// ))
//
// This method is thread-safe.
func (p *Pools) Connect(spec db.Specifier, err error) (db.DB, error) {
	// Handle any error passed in by the db.Config
	if err != nil {
		return nil, err
	}

	pool, err := p.poolForSpec(spec)
	if err != nil {
		return nil, err
	}

	return Connect(pool, spec)
}

// Connect checks whether a connection has been created with this unique
// specifier.  If so, it checks whether it needs to be updated, and returns it.
// Otherwise, it will create the new connection.
func Connect(pool ServerPool, spec db.Specifier) (db.DB, error) {
	pool.RLock()

	if db, ok := pool.Get(spec); ok {
		// The connection was found.  Do we need to update it?
		if db.Spec().NeedsUpdate(spec) {
			pool.RUnlock()
			pool.Lock()
			// By now, someone else may have updated it already.
			if db.Spec().NeedsUpdate(spec) {
				if err := db.Update(spec); err != nil {
					pool.Unlock()
					return nil, err
				}
			}
			pool.Unlock()
			pool.RLock()
		}
		pool.RUnlock()
		return db, nil
	}
	pool.RUnlock()
	return insertNewDB(pool, spec)
}

// newDB adds a new DB generated using the Specifier's newDB method to the pool
// and returns it.
func insertNewDB(pool ServerPool, spec db.Specifier) (db.DB, error) {
	// If we need to create the db connection, then we need to write-lock
	// the DB pool.
	pool.Lock()
	defer pool.Unlock()

	if db, ok := pool.Get(spec); ok {
		// Someone else may have already come along and added it
		if db.Spec().NeedsUpdate(spec) {
			if err := db.Update(spec); err != nil {
				return nil, err
			}
		}
		return db, nil
	}

	newDb, err := spec.NewDB()
	if err != nil {
		return nil, err
	}

	pool.Put(spec, newDb)
	newDb, ok := pool.Get(spec)
	if !ok {
		return nil, fmt.Errorf("new database not found for %T", spec)
	}
	return newDb, nil
}

// FlushEntry can be used to flush an entry from a correct pool.  Note that if a
// handle to the entry still exists, it will remain in memory until released.
func FlushEntry(pool ServerPool, spec db.Specifier) {
	pool.Delete(spec)
}
