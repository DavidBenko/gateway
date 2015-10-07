package sql_test

import (
	"encoding/json"

	sql "gateway/db/sql"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func pqConfigs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"simple": map[string]interface{}{
			"host":     "some.url.net",
			"port":     1234,
			"user":     "user",
			"password": "pass",
			"dbname":   "db",
			"sslmode":  "prefer",
		},
		"bad": map[string]interface{}{
			"port":   1234,
			"user":   "user",
			"dbname": "db",
		},
		"badport": map[string]interface{}{
			"port":     -1234,
			"user":     "user",
			"password": "pass",
			"dbname":   "db",
		},
		"complicated": map[string]interface{}{
			"host":     "some.url.net",
			"port":     1234,
			"user":     "user name",
			"password": "pass's",
			"dbname":   "db",
			"sslmode":  "verify-ca",
		},
	}
}

func pqSpecs() map[string]*sql.PostgresSpec {
	specs := make(map[string]*sql.PostgresSpec)
	for name, spc := range pqConfigs() {
		pqSpec := &sql.PostgresSpec{}
		vals, err := json.Marshal(spc)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(vals, pqSpec)

		specs[name] = pqSpec
	}
	return specs
}

func (s *SQLSuite) TestPostgresConfig(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        *sql.PostgresSpec
		expectString string
		expectUnique string
		expectError  string
	}{{
		should:       "work with a simple config",
		given:        pqSpecs()["simple"],
		expectString: "postgres://user:pass@some.url.net:1234/db?sslmode=prefer",
		expectUnique: "postgres://user:pass@some.url.net:1234/db?sslmode=prefer",
	}, {
		should:       "work with a complicated config",
		given:        pqSpecs()["complicated"],
		expectString: `postgres://user name:pass's@some.url.net:1234/db?sslmode=verify-ca`,
		expectUnique: `postgres://user name:pass's@some.url.net:1234/db?sslmode=verify-ca`,
	}, {
		should:      "not work with a bad config",
		given:       pqSpecs()["bad"],
		expectError: `pgx config errors: bad value "" for "password"; bad value "" for "host"`,
	}, {
		should:      "not work with a bad config",
		given:       pqSpecs()["badport"],
		expectError: `pgx config errors: bad value -1234 for "port"; bad value "" for "host"`,
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
