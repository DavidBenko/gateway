package redis_test

import (
	"encoding/json"
	"fmt"
	"gateway/db"
	"gateway/db/mongo"
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
		expectError: "redis config errors: requires Host; requires Port",
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

func (r *RedisSuite) TestNeedsUpdate(c *gc.C) {
	self := redisSpecs()["simple"]

	// Redis spec containing a different pool limit
	poolLimitSpec, err := redis.Config(redis.Connection(redisSpecs()["simple"]), redis.MaxActive(5))
	if err != nil {
		panic(err)
	}

	for i, t := range []struct {
		should      string
		given       db.Specifier
		compare     db.Specifier
		expect      bool
		expectPanic string
	}{{
		should:  "not error a self-check",
		given:   self,
		compare: self,
	}, {
		should:      "not work comparing different types",
		given:       redisSpecs()["simple"],
		compare:     &mongo.Spec{},
		expectPanic: "tried to compare wrong database kinds: Redis and *mongo.Spec",
	}, {
		should:      "fail to compare to nil",
		given:       redisSpecs()["simple"],
		expectPanic: "tried to compare to nil db.Specifier!",
	}, {
		should:  "return true if supplied a different limit",
		given:   redisSpecs()["simple"],
		compare: poolLimitSpec,
		expect:  true,
	}} {
		message := fmt.Sprintf("Test %d: should %s", i, t.should)
		if t.expectPanic != "" {
			message += " (expect panic)"
		}

		c.Log(message)

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

			result := t.given.NeedsUpdate(t.compare)
			c.Check(result, gc.Equals, t.expect)

		}()
	}
}

func (r *RedisSuite) TestUpdate(c *gc.C) {
	for i, t := range []struct {
		should       string
		givenLimit   int
		newLimit     int
		expectUpdate bool
	}{{
		should:       "not update if limit is the same",
		givenLimit:   0,
		newLimit:     0,
		expectUpdate: false,
	}, {
		should:       "update if limit changes",
		givenLimit:   0,
		newLimit:     5,
		expectUpdate: true,
	}} {
		c.Logf("Test %d: should %s", i, t.should)
		conf, err := redis.Config(
			redis.Connection(redisSpecs()["simple"]),
			redis.MaxActive(t.givenLimit),
		)
		c.Assert(err, gc.IsNil)
		newConf, err := redis.Config(
			redis.Connection(redisSpecs()["simple"]),
			redis.MaxActive(t.newLimit),
		)
		c.Assert(err, gc.IsNil)
		c.Logf("Test %d:\n old: %v\n new: %+v\n", i, conf, newConf)
		c.Check(conf.NeedsUpdate(newConf), gc.Equals, t.expectUpdate)
	}
}
