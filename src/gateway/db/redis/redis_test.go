package redis_test

import (
	"encoding/json"
	"gateway/db/redis"
	"testing"

	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type RedisSuite struct{}

var _ = gc.Suite(&RedisSuite{})

func redisConfigs() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"simple": map[string]interface{}{
			"username": "admin",
			"password": "password",
			"host":     "redisserver",
			"port":     6379,
			"database": "test",
		},
		"bad": map[string]interface{}{
			"username": "",
		},
	}
}

func redisSpecs() map[string]*redis.Spec {
	specs := make(map[string]*redis.Spec)
	for name, spc := range redisConfigs() {
		redisSpec := &redis.Spec{}
		vals, err := json.Marshal(spc)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(vals, redisSpec)
		if err != nil {
			panic(err)
		}
		specs[name] = redisSpec
	}
	return specs
}

func (r *RedisSuite) TestRedisConfig(c *gc.C) {
	for i, t := range []struct {
		should       string
		given        *redis.Spec
		expectString string
		expectUnique string
		expectError  string
	}{{
		should:       "work with a simple config",
		given:        redisSpecs()["simple"],
		expectString: "redis://admin:password@redisserver:6379/test",
		expectUnique: "redis://admin:password@redisserver:6379/test",
	}, {
		should:      "not work with a bad config",
		given:       redisSpecs()["bad"],
		expectError: "redis config errors: requires Username; requires Password; requires Host; requires Port",
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		conf, err := redis.Config(redis.Connection(t.given))

		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}

		c.Assert(err, gc.IsNil)
		c.Check(conf.ConnectionString(), gc.Equals, t.expectString)
		c.Check(conf.UniqueServer(), gc.Equals, t.expectUnique)

	}
}
