package admin_test

import (
	"os"
	"testing"

	"gateway/config"
	apsql "gateway/sql"

	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type AdminSuite struct {
	db *apsql.DB
}

var _ = gc.Suite(&AdminSuite{})

func newDB(c *gc.C, conf config.Database) *apsql.DB {
	c.Logf("connecting to database %v", conf)
	db, err := apsql.Connect(conf)
	c.Assert(err, gc.IsNil)
	c.Assert(db.Migrate(), gc.IsNil)

	return db
}

func (m *AdminSuite) SetUpTest(c *gc.C) {
	if db := m.db; db != nil {
		c.Assert(db.Close(), gc.IsNil)
	}

	m.db = newDB(c, config.Database{
		Driver:           "sqlite3",
		ConnectionString: "/tmp/stats",
	})
}

func (m *AdminSuite) TearDownTest(c *gc.C) {
	if db := m.db; db != nil {
		c.Assert(db.Close(), gc.IsNil)
	}
	os.Remove("/tmp/stats")
}
