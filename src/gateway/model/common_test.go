package model_test

import (
	"testing"

	"gateway/config"
	modelt "gateway/model/testing"
	apsql "gateway/sql"

	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type ModelSuite struct {
	db *apsql.DB
}

var _ = gc.Suite(&ModelSuite{})

func (m *ModelSuite) SetUpTest(c *gc.C) {
	if db := m.db; db != nil {
		c.Assert(db.Close(), gc.IsNil)
	}

	m.db = modelt.NewDB(c, config.Database{
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
	})
}

func (m *ModelSuite) TearDownTest(c *gc.C) {
	if db := m.db; db != nil {
		c.Assert(db.Close(), gc.IsNil)
	}
}
