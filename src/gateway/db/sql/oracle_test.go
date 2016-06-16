package sql_test

import (
	"encoding/json"

	sql "gateway/db/sql"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func OracleConfigs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"simple": map[string]interface{}{
			"host":     "localhost",
			"port":     1521,
			"user":     "system",
			"password": "manager",
			"dbname":   "orcl",
		},
		"bad": map[string]interface{}{
			"port":   -1234,
			"user":   "user",
			"dbname": "db",
		},
	}
}

func oracleSpecs() map[string]*sql.OracleSpec {
	specs := make(map[string]*sql.OracleSpec)
	for name, spc := range OracleConfigs() {
		oracleSpec := &sql.OracleSpec{}
		vals, err := json.Marshal(spc)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(vals, oracleSpec)

		specs[name] = oracleSpec
	}
	return specs
}

func (s *SQLSuite) TestOracleConfig(c *gc.C) {

	for i, t := range []struct {
		should       string
		given        *sql.OracleSpec
		expectString string
		expectUnique string
		expectError  string
	}{{
		should:       "work with a simple config",
		given:        oracleSpecs()["simple"],
		expectString: "oci8://system:manager@localhost:1521/orcl",
		expectUnique: "oci8://system:manager@localhost:1521/orcl",
	}, {
		should: "fail with multiple bad config items",
		given:  oracleSpecs()["bad"],
		expectError: `oci8 config errors: ` +
			`bad value -1234 for "port"; ` +
			`bad value "" for "password"; ` +
			`bad value "" for "host"`,
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
