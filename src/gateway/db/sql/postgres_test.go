package sql_test

import (
	"encoding/json"
	"testing"

	sql "gateway/db/sql"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type PostgresSuite struct{}

var _ = gc.Suite(&PostgresSuite{})

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

func (s *PostgresSuite) TestPostgresConfig(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        map[string]interface{}
		expectString string
		expectUnique string
		expectError  string
	}{{
		should:       "work with a simple config",
		given:        pqConfigs()["simple"],
		expectString: "postgres://user:pass@some.url.net:1234/db?sslmode=prefer",
		expectUnique: "postgres://user:pass@some.url.net:1234/db?sslmode=prefer",
	}, {
		should:       "work with a complicated config",
		given:        pqConfigs()["complicated"],
		expectString: `postgres://user name:pass's@some.url.net:1234/db?sslmode=verify-ca`,
		expectUnique: `postgres://user name:pass's@some.url.net:1234/db?sslmode=verify-ca`,
	}, {
		should:      "not work with a bad config",
		given:       pqConfigs()["bad"],
		expectError: `pgx config errors: bad value "" for "password"; bad value "" for "host"`,
	}, {
		should:      "not work with a bad config",
		given:       pqConfigs()["badport"],
		expectError: `pgx config errors: bad value -1234 for "port"; bad value "" for "host"`,
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		pqConn := &sql.PostgresSpec{}
		vals, err := json.Marshal(t.given)
		c.Assert(err, jc.ErrorIsNil)
		c.Logf("Test %d:\n  given: %s", i, vals)
		err = json.Unmarshal(vals, pqConn)
		c.Assert(err, jc.ErrorIsNil)

		obtained, err := sql.Config(sql.Connection(pqConn))
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}
		c.Assert(err, jc.ErrorIsNil)
		c.Check(obtained.ConnectionString(), gc.Equals, t.expectString)
		c.Check(obtained.UniqueServer(), gc.Equals, t.expectUnique)
	}
}
