package sql_test

import (
	"encoding/json"
	"fmt"

	"gateway/db"
	sql "gateway/db/sql"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func mysqlConfigs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"simple": map[string]interface{}{
			"server":   "some.url.net",
			"port":     1234,
			"username": "username",
			"password": "pass",
			"dbname":   "db",
		},
		"complicated": map[string]interface{}{
			"server":   "some.url.net",
			"port":     1234,
			"username": "user name",
			"password": "pass's",
			"dbname":   "db",
			"timeout":  "30s",
		},
		"bad": map[string]interface{}{
			"port":     1234,
			"username": "username",
			"dbname":   "db",
		},
		"badport": map[string]interface{}{
			"port":     -1234,
			"username": "username",
			"password": "pass",
			"dbname":   "db",
		},
	}
}

func mysqlSpecs() map[string]*sql.MySQLSpec {
	specs := make(map[string]*sql.MySQLSpec)
	for name, spc := range mysqlConfigs() {
		mysqlSpec := &sql.MySQLSpec{}
		vals, err := json.Marshal(spc)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(vals, mysqlSpec)
		if err != nil {
			panic(err)
		}

		specs[name] = mysqlSpec
	}
	return specs
}

func (s *SQLSuite) TestMySQLConfig(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        *sql.MySQLSpec
		expectString string
		expectUnique string
		expectError  string
	}{{
		should:       "work with a simple config",
		given:        mysqlSpecs()["simple"],
		expectString: "username:pass@tcp(some.url.net:1234)/db",
		expectUnique: "username:pass@tcp(some.url.net:1234)/db",
	}, {
		should:       "work with a complicated config",
		given:        mysqlSpecs()["complicated"],
		expectString: `user name:pass's@tcp(some.url.net:1234)/db?timeout=30s`,
		expectUnique: `user name:pass's@tcp(some.url.net:1234)/db`,
	}, {
		should:      "not work with a bad config",
		given:       mysqlSpecs()["bad"],
		expectError: `mysql config errors: bad value "" for "password"; bad value "" for "server"`,
	}, {
		should:      "not work with a nil config",
		expectError: `cannot create SQL Connection with nil Specifier`,
	}, {
		should:      "not work with a bad config",
		given:       mysqlSpecs()["badport"],
		expectError: `mysql config errors: bad value -1234 for "port"; bad value "" for "server"`,
	}} {
		c.Logf("Test %d: should %s", i, t.should)

		obtained, err := sql.Config(sql.Connection(t.given))
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}
		c.Assert(err, jc.ErrorIsNil)
		c.Check(obtained.ConnectionString(), gc.Equals, t.expectString)
		c.Check(obtained.UniqueServer(), gc.Equals, t.expectUnique)
	}
}

func (s *SQLSuite) TestMySQLNeedsUpdate(c *gc.C) {
	self := mysqlSpecs()["simple"]
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
		given:   mysqlSpecs()["simple"],
		compare: mysqlSpecs()["complicated"],
		expect:  true,
	}, {
		should:      "not work comparing different types",
		given:       mysqlSpecs()["simple"],
		compare:     sqlsSpecs()["simple"],
		expectPanic: "tried to compare *sql.SQLServerSpec to *sql.MySQLSpec!",
	}, {
		should:      "fail to compare nil specs",
		given:       mysqlSpecs()["simple"],
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
