package db

// DB wraps a typed db handle.  spec returns the db's
// Specifier, while update updates the db's config options
// (e.g. changing max. idle dbs.)
type DB interface {
	Spec() Specifier
	Update(Specifier) error
}

// Specifier defines methods that a DB connection specifier must implement for
// use in a pool.
type Specifier interface {
	ConnectionString() string
	UniqueServer() string
	NeedsUpdate(Specifier) bool
	NewDB() (DB, error)
}

// Config sets up a Specifier and should be implemented per database, as in:
//
// import sqls "gateway/db/sqlserver"
//
// pools.Connect(db.SQLConfig(
//         sqls.Connection("hello.com", 4445),
//         sqls.MaxOpenIdle(10, 100),
// ))
//
// where DB and MaxIdle are of type Config.
type Config func(...Configurator) (Specifier, error)

// configurator is a function for updating a Specifier.
type Configurator func(Specifier) error
