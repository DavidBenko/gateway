package mongo_test

import (
	"testing"

	mongo "gateway/db/mongo"

	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type MongoSuite struct{}

var _ = gc.Suite(&MongoSuite{})

func configs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"simple": map[string]interface{}{
			"hosts": []interface{}{
				map[string]interface{}{
					"host": "test.com",
					"port": float64(123),
				},
			},
			"username": "user",
			"password": "pass",
			"database": "db",
		},
		"bad": map[string]interface{}{
			"hosts": []interface{}{
				map[string]interface{}{
					"host": "test.com",
				},
			},
			"username": "user",
			"password": "pass",
			"database": "db",
		},
		"bad2": map[string]interface{}{
			"hosts":    nil,
			"username": "user",
			"password": "pass",
			"database": "db",
		},
		"bad3": map[string]interface{}{
			"hosts": []interface{}{
				map[string]interface{}{
					"port": float64(123),
				},
			},
			"username": "user",
			"password": "pass",
			"database": "db",
		},
		"complicated": map[string]interface{}{
			"hosts": []interface{}{
				map[string]interface{}{
					"host": "test.com",
					"port": float64(123),
				}, map[string]interface{}{
					"host": "another.com",
					"port": float64(123),
				},
			},
			"username": "user name",
			"password": "pass's",
			"database": "db",
		},
	}
}

func (s *MongoSuite) TestMongoConfig(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        map[string]interface{}
		expectString string
		expectUnique string
		expectError  string
	}{{
		should:       "work with a simple config",
		given:        configs()["simple"],
		expectString: "mongodb://user:pass@test.com:123/db",
		expectUnique: "mongodb://user:pass@test.com:123/db",
	}, {
		should:       "work with a complicated config",
		given:        configs()["complicated"],
		expectString: `mongodb://user+name:pass%27s@test.com:123,another.com:123/db`,
		expectUnique: `mongodb://user+name:pass%27s@test.com:123,another.com:123/db`,
	}, {
		should:      "not work with a bad config",
		given:       configs()["bad"],
		expectError: `Port is required`,
	}, {
		should:      "not work with a bad config",
		given:       configs()["bad2"],
		expectError: `At least one host must be defined.`,
	}, {
		should:      "not work with a bad config",
		given:       configs()["bad3"],
		expectError: `Host name is required`,
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		conf, err := mongo.Config(mongo.Connection(t.given))
		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}
		c.Assert(err, gc.IsNil)
		c.Check(conf.ConnectionString(), gc.Equals, t.expectString)
		c.Check(conf.UniqueServer(), gc.Equals, t.expectUnique)
	}
}

func (s *MongoSuite) TestMongoUpdate(c *gc.C) {
	for i, t := range []struct {
		should       string
		givenLimit   int
		newLimit     int
		expectUpdate bool
	}{{
		should:       "not expect update if limit remains the same",
		givenLimit:   0,
		newLimit:     0,
		expectUpdate: false,
	}, {
		should:       "expect update if limit changes",
		givenLimit:   0,
		newLimit:     100,
		expectUpdate: true,
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		conf, err := mongo.Config(
			mongo.Connection(configs()["simple"]),
			mongo.PoolLimit(t.givenLimit),
		)
		c.Assert(err, gc.IsNil)
		newConf, err := mongo.Config(
			mongo.Connection(configs()["simple"]),
			mongo.PoolLimit(t.newLimit),
		)
		c.Assert(err, gc.IsNil)
		c.Logf("Test %d:\n  old: %+v\n  new: %+v\n", i, conf, newConf)
		c.Check(conf.NeedsUpdate(newConf), gc.Equals, t.expectUpdate)
	}
}
