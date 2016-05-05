package sql_test

import (
	"encoding/json"
	"fmt"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"gateway/db"
	sql "gateway/db/sql"
)

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
			"schema":             "complexSchema",
			"connection timeout": 30,
			"encrypt":            "true",
		},
	}
}

func sqlsSpecs() map[string]*sql.SQLServerSpec {
	sqlsSpecs := make(map[string]*sql.SQLServerSpec)
	for name, spc := range sqlsConfigs() {
		sqlsSpec := &sql.SQLServerSpec{}
		vals, err := json.Marshal(spc)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(vals, sqlsSpec)
		if err != nil {
			panic(err)
		}
		sqlsSpecs[name] = sqlsSpec
	}
	return sqlsSpecs
}

func (s *SQLSuite) TestSQLSConfig(c *gc.C) {
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
		expectString: "database=db;encrypt=true;password=pass;port=1234;schema=complexSchema;server=some.url.net;timeout=30;user id=user",
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

func (s *SQLSuite) TestSQLServerNeedsUpdate(c *gc.C) {
	self := sqlsSpecs()["simple"]
	for i, t := range []struct {
		should      string
		given       db.Specifier
		compare     db.Specifier
		expect      bool
		expectPanic string
	}{{
		should:  "not error on a self-check",
		given:   self,
		compare: self,
	}, {
		should:  "be true if needs update",
		given:   sqlsSpecs()["simple"],
		compare: sqlsSpecs()["complicated"],
		expect:  true,
	}, {
		should:      "not work comparing different types",
		given:       sqlsSpecs()["simple"],
		compare:     mysqlSpecs()["simple"],
		expectPanic: "tried to compare wrong database kinds: *sql.MySQLSpec and *sql.SQLServerSpec",
	}, {
		should:      "fail to compare nil specs",
		given:       sqlsSpecs()["simple"],
		expectPanic: "tried to compare to nil db.Specifier!",
	}} {
		msg := fmt.Sprintf("Test %d: should %s", i, t.should)
		if t.expectPanic != "" {
			msg += " (expect panic)"
		}

		c.Logf(msg)

		func() {
			defer func() {
				e := recover()
				switch {
				case t.expectPanic != "":
					c.Assert(e, gc.Equals, t.expectPanic)
				default:
					c.Assert(e, gc.IsNil)
				}
			}()

			c1, c2 := t.given, t.compare
			switch {
			case c1 == nil && c2 == nil:
				c.Log("tried to compare a nil spec to a nil spec")
				c.FailNow()
			case c1 == nil:
				result := c2.NeedsUpdate(c1)
				c.Check(result, gc.Equals, t.expect)
			case c2 == nil:
				result := c1.NeedsUpdate(c2)
				c.Check(result, gc.Equals, t.expect)
			default:
				result := c2.NeedsUpdate(c1)
				c.Check(result, gc.Equals, t.expect)
				result = c1.NeedsUpdate(c2)
				c.Check(result, gc.Equals, t.expect)
			}
		}()
	}
}
