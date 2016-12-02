package testing

import (
	"gateway/config"
	apsql "gateway/sql"

	gc "gopkg.in/check.v1"
)

// NewDB gets a new database handle given a config.
func NewDB(c *gc.C, conf config.Database) *apsql.DB {
	c.Logf("connecting to database %v", conf)
	db, err := apsql.Connect(conf)
	c.Assert(err, gc.IsNil)
	c.Assert(db.Migrate(), gc.IsNil)

	return db
}
