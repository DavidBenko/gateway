package postgres_test

import (
	"testing"

	psql "gateway/db/postgres"

	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type PostgresSuite struct{}

var _ = gc.Suite(&PostgresSuite{})

func configs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"simple": map[string]interface{}{
			"host":     "some.url.net",
			"port":     1234,
			"user":     "user",
			"password": "pass",
			"dbname":   "db",
		},
		"bad": map[string]interface{}{
			"port":     1234,
			"user":     "user",
			"password": "pass",
			"dbname":   "db",
		},
		"complicated": map[string]interface{}{
			"host":            "some.url.net",
			"port":            1234,
			"user":            "user name",
			"password":        "pass's",
			"dbname":          "db",
			"connect_timeout": 30,
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
		given:        configs()["simple"],
		expectString: "dbname=db host=some.url.net password=pass port=1234 user=user",
		expectUnique: "dbname=db host=some.url.net password=pass port=1234 user=user",
	}, {
		should:       "work with a complicated config",
		given:        configs()["complicated"],
		expectString: `connect_timeout=30 dbname=db host=some.url.net password=pass\'s port=1234 user='user name'`,
		expectUnique: `dbname=db host=some.url.net password=pass\'s port=1234 user='user name'`,
	}, {
		should:      "not work with a bad config",
		given:       configs()["bad"],
		expectError: `Postgres Config missing "host" key`,
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		conf, err := psql.Config(psql.Connection(t.given))
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}
		c.Assert(err, gc.IsNil)
		c.Check(conf.ConnectionString(), gc.Equals, t.expectString)
		c.Check(conf.UniqueServer(), gc.Equals, t.expectUnique)
	}
}

func (s *PostgresSuite) TestPostgresUpdate(c *gc.C) {
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
		conf, err := psql.Config(
			psql.Connection(configs()["simple"]),
			psql.MaxOpenIdle(t.givenMaxOpen, t.givenMaxIdle),
		)
		c.Assert(err, gc.IsNil)
		newConf, err := psql.Config(
			psql.Connection(configs()["simple"]),
			psql.MaxOpenIdle(t.newMaxOpen, t.newMaxIdle),
		)
		c.Assert(err, gc.IsNil)
		c.Logf("Test %d:\n  old: %+v\n  new: %+v\n", i, conf, newConf)
		c.Check(conf.NeedsUpdate(newConf), gc.Equals, t.expectUpdate)
	}
}
