package stats_test

import (
	"testing"

	"gateway/errors"
	"gateway/stats"

	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type StatsSuite struct{}

var _ = gc.Suite(&StatsSuite{})

func (s *StatsSuite) TestKeyValidate(c *gc.C) {
	for i, test := range []struct {
		given  string
		expect errors.Errors
	}{
		{given: "proxy.env.id", expect: errors.Errors{}},
		{given: "proxy.env.name", expect: errors.Errors{}},
		{given: "proxy.group.id", expect: errors.Errors{}},
		{given: "proxy.group.name", expect: errors.Errors{}},
		{given: "proxy.id", expect: errors.Errors{}},
		{given: "proxy.name", expect: errors.Errors{}},
		{given: "proxy.route.path", expect: errors.Errors{}},
		{given: "proxy.route.verb", expect: errors.Errors{}},
		{given: "remote_endpoint.response.time", expect: errors.Errors{}},
		{given: "request.size", expect: errors.Errors{}},
		{given: "request.id", expect: errors.Errors{}},
		{given: "response.time", expect: errors.Errors{}},
		{given: "response.error", expect: errors.Errors{}},
		{given: "response.size", expect: errors.Errors{}},
		{given: "response.status", expect: errors.Errors{}},
		{given: "api.id", expect: errors.Errors{}},
		{given: "api.name", expect: errors.Errors{}},
		{given: "host.id", expect: errors.Errors{}},
		{given: "host.name", expect: errors.Errors{}},
		{given: "node", expect: errors.Errors{}},
		{given: "timestamp", expect: errors.Errors{}},
		{given: "remote.endpoint.response.time", expect: errors.Errors{
			"key": {`invalid measurement "remote.endpoint.response.time"`},
		}},
		{given: "ms", expect: errors.Errors{
			"key": {`invalid measurement "ms"`},
		}},
		{given: "request_id", expect: errors.Errors{
			"key": {`invalid measurement "request_id"`},
		}},
		{given: "request.time", expect: errors.Errors{
			"key": {`invalid measurement "request.time"`},
		}},
		{given: "ho ho", expect: errors.Errors{
			"key": {`invalid measurement "ho ho"`},
		}},
	} {
		c.Logf("test %d: %v validation is %#q", i, test.given, test.expect)
		given := stats.Constraint{Key: test.given, Operator: stats.EQ}
		c.Check(given.Validate(), gc.DeepEquals, test.expect)
	}
}

func (s *StatsSuite) TestOperatorValidate(c *gc.C) {
	for i, test := range []struct {
		given  stats.Operator
		expect errors.Errors
	}{
		{given: stats.EQ, expect: errors.Errors{}},
		{given: stats.GT, expect: errors.Errors{}},
		{given: stats.GTE, expect: errors.Errors{}},
		{given: stats.LT, expect: errors.Errors{}},
		{given: stats.LTE, expect: errors.Errors{}},
		{given: stats.IN, expect: errors.Errors{}},
		{given: stats.Operator("ho ho"), expect: errors.Errors{
			"operator": {`invalid operator "ho ho"`},
		}},
		{given: stats.Operator("<<"), expect: errors.Errors{
			"operator": {`invalid operator "<<"`},
		}},
	} {
		c.Logf("test %d: %v validation is %#q", i, test.given, test.expect)
		given := stats.Constraint{Key: "node", Operator: test.given}
		c.Check(given.Validate(), gc.DeepEquals, test.expect)
	}
}

func (s *StatsSuite) TestAllMeasurements(c *gc.C) {
	c.Assert(stats.AllMeasurements(), gc.DeepEquals, []string{
		`api.id`,
		`api.name`,
		`host.id`,
		`host.name`,
		`proxy.env.id`,
		`proxy.env.name`,
		`proxy.group.id`,
		`proxy.group.name`,
		`proxy.id`,
		`proxy.name`,
		`proxy.route.path`,
		`proxy.route.verb`,
		`remote_endpoint.response.time`,
		`request.id`,
		`request.size`,
		`response.error`,
		`response.size`,
		`response.status`,
		`response.time`,
	})
}

func (s *StatsSuite) TestAllSamples(c *gc.C) {
	c.Assert(stats.AllSamples(), gc.DeepEquals, []string{
		`api.id`,
		`api.name`,
		`host.id`,
		`host.name`,
		`node`,
		`proxy.env.id`,
		`proxy.env.name`,
		`proxy.group.id`,
		`proxy.group.name`,
		`proxy.id`,
		`proxy.name`,
		`proxy.route.path`,
		`proxy.route.verb`,
		`remote_endpoint.response.time`,
		`request.id`,
		`request.size`,
		`response.error`,
		`response.size`,
		`response.status`,
		`response.time`,
		`timestamp`,
	})
}
