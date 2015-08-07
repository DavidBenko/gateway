package sqlserver_test

import (
	"testing"

	gc "gopkg.in/check.v1"

	sqls "gateway/db/sqlserver"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type SQLServerSuite struct{}

var _ = gc.Suite(&SQLServerSuite{})

func configs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"simple": map[string]interface{}{
			"server":   "some.url.net",
			"port":     1234,
			"user id":  "user",
			"password": "pass",
			"database": "db",
			"schema":   "dbschema",
		},
		"complicated": map[string]interface{}{
			"server":             "some.url.net",
			"port":               1234,
			"user id":            "user",
			"password":           "pass",
			"database":           "db",
			"schema":             "dbschema",
			"connection timeout": 30,
		},
	}
}

func (s *SQLServerSuite) TestSQLConfig(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        map[string]interface{}
		expectString string
		expectUnique string
	}{{
		should:       "work with a simple config",
		given:        configs()["simple"],
		expectString: "database=db;password=pass;port=1234;schema=dbschema;server=some.url.net;user id=user;",
		expectUnique: "database=db;password=pass;port=1234;schema=dbschema;server=some.url.net;user id=user;",
	}, {
		should:       "work with a complicated config",
		given:        configs()["complicated"],
		expectString: "connection timeout=30;database=db;password=pass;port=1234;schema=dbschema;server=some.url.net;user id=user;",
		expectUnique: "database=db;password=pass;port=1234;schema=dbschema;server=some.url.net;user id=user;",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		conf, err := sqls.Config(sqls.Connection(t.given))
		c.Assert(err, gc.IsNil)
		c.Check(conf.ConnectionString(), gc.Equals, t.expectString)
		c.Check(conf.UniqueServer(), gc.Equals, t.expectUnique)
	}
}

func (s *SQLServerSuite) TestSQLServerUpdate(c *gc.C) {
	for i, t := range []struct {
		should       string
		givenMaxOpen int
		givenMaxIdle int
		newMaxOpen   int
		newMaxIdle   int
		expectUpdate bool
	}{{
		should:       "not expect update if max open and max idle remain the same",
		givenMaxOpen: 0,
		givenMaxIdle: 0,
		newMaxOpen:   0,
		newMaxIdle:   0,
		expectUpdate: false,
	}, {
		should:       "expect update if only max open changes",
		givenMaxOpen: 0,
		givenMaxIdle: 0,
		newMaxOpen:   10,
		newMaxIdle:   0,
		expectUpdate: true,
	}, {
		should:       "expect update if only max idle changes",
		givenMaxOpen: 0,
		givenMaxIdle: 0,
		newMaxOpen:   0,
		newMaxIdle:   10,
		expectUpdate: true,
	}, {
		should:       "expect update if both max idle and max open change",
		givenMaxOpen: 0,
		givenMaxIdle: 0,
		newMaxOpen:   10,
		newMaxIdle:   10,
		expectUpdate: true,
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		conf, err := sqls.Config(
			sqls.Connection(configs()["simple"]),
			sqls.MaxOpenIdle(t.givenMaxOpen, t.givenMaxIdle),
		)
		c.Assert(err, gc.IsNil)
		newConf, err := sqls.Config(
			sqls.Connection(configs()["simple"]),
			sqls.MaxOpenIdle(t.newMaxOpen, t.newMaxIdle),
		)
		c.Assert(err, gc.IsNil)
		c.Logf("Test %d:\n  old: %+v\n  new: %+v\n", i, conf, newConf)
		c.Check(conf.NeedsUpdate(newConf), gc.Equals, t.expectUpdate)
	}
}
