package sql_test

import (
	"encoding/json"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	sql "gateway/db/sql"
)

type SQLServerSuite struct{}

var _ = gc.Suite(&SQLServerSuite{})

func sqlsConfigs() map[string]map[string]interface{} {
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
		expectError  string
	}{{
		should:       "work with a simple config",
		given:        sqlsConfigs()["simple"],
		expectString: "database=db;password=pass;port=1234;schema=dbschema;server=some.url.net;user id=user",
		expectUnique: "dbname=db;host=some.url.net;password=pass;port=1234;user id=user",
	}, {
		should:       "work with a complicated config",
		given:        sqlsConfigs()["complicated"],
		expectString: "database=db;password=pass;port=1234;schema=dbschema;server=some.url.net;timeout=30;user id=user",
		expectUnique: "dbname=db;host=some.url.net;password=pass;port=1234;user id=user",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		sqlsConf := &sql.SQLServerSpec{}
		vals, err := json.Marshal(t.given)
		c.Assert(err, jc.ErrorIsNil)
		err = json.Unmarshal(vals, sqlsConf)
		c.Assert(err, jc.ErrorIsNil)

		obtained, err := sql.Config(sql.Connection(sqlsConf))
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}
		c.Assert(err, jc.ErrorIsNil)
		c.Check(obtained.ConnectionString(), gc.Equals, t.expectString)
		c.Check(obtained.UniqueServer(), gc.Equals, t.expectUnique)
	}
}
