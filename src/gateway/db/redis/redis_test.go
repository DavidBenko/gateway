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
			"user":     "admin",
			"password": "password",
			"host":     "redisserver",
			"port":     6379,
			"database": "test",
		},
		"bad": map[string]interface{}{
			"user": "admin",
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
	}} {
		c.Logf("Test %d: should %s", i, t.should)

	}
}
