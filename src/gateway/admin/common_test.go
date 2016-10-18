package admin_test

import (
	"testing"

	"gateway/config"
	apsql "gateway/sql"
	statssql "gateway/stats/sql"

	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type AdminSuite struct {
	db      *apsql.DB
	statsDb *statssql.SQL
}

var _ = gc.Suite(&AdminSuite{})

func newDB(c *gc.C, conf config.Database) *apsql.DB {
	c.Logf("connecting to database %v", conf)
	db, err := apsql.Connect(conf)
	c.Assert(err, gc.IsNil)
	c.Assert(db.Migrate(), gc.IsNil)

	return db
}

func newStatsDB(c *gc.C, conf config.Stats) *statssql.SQL {
	c.Logf("connecting to stats database %v", conf)
	db, err := statssql.Connect(conf)
	c.Assert(err, gc.IsNil)
	c.Assert(db.Migrate(), gc.IsNil)

	return db
}

func (m *AdminSuite) SetUpTest(c *gc.C) {
	m.db = newDB(c, config.Database{
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
	})
	m.statsDb = newStatsDB(c, config.Stats{
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
	})
}

func (m *AdminSuite) TearDownTest(c *gc.C) {
	if db := m.db; db != nil {
		c.Assert(db.Close(), gc.IsNil)
	}
	if statsDb := m.statsDb; statsDb != nil {
		c.Assert(statsDb.Close(), gc.IsNil)
	}
}
