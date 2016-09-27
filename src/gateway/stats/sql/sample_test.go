package sql_test

import (
	"time"

	"gateway/stats"
	"gateway/stats/sql"

	"github.com/jmoiron/sqlx"
	gc "gopkg.in/check.v1"
)

func (s *SQLSuite) TestSampleQuery(c *gc.C) {
	for i, test := range []struct {
		should           string
		givenConstraints []stats.Constraint
		givenVars        []string
		givenDriver      *sqlx.DB
		expect           string
	}{{
		should:      "generate a minimal sampler query for SQLite",
		givenVars:   []string{"request.size", "request.id"},
		givenDriver: s.sqlite,
		expect: `
SELECT
  request_size
  , request_id
FROM stats
ORDER BY timestamp, node`[1:],
	}, {
		should:      "generate a minimal sampler query for Postgres",
		givenVars:   []string{"request.size", "request.id"},
		givenDriver: s.postgres,
		expect: `
SELECT
  request_size
  , request_id
FROM stats
ORDER BY timestamp, node`[1:],
	}, {
		should: "generate a simple sampler query for SQLite",
		givenConstraints: []stats.Constraint{
			{Key: "timestamp", Operator: stats.GTE, Value: time.Now()},
			{Key: "timestamp", Operator: stats.LT, Value: time.Now()},
		},
		givenVars:   []string{"request.size", "request.id"},
		givenDriver: s.sqlite,
		expect: `
SELECT
  request_size
  , request_id
FROM stats
WHERE timestamp >= ?
  AND timestamp < ?
ORDER BY timestamp, node`[1:],
	}, {
		should: "generate a simple sampler query for Postgres",
		givenConstraints: []stats.Constraint{
			{Key: "timestamp", Operator: stats.GTE, Value: time.Now()},
			{Key: "timestamp", Operator: stats.LT, Value: time.Now()},
		},
		givenVars:   []string{"request.size", "request.id"},
		givenDriver: s.postgres,
		expect: `
SELECT
  request_size
  , request_id
FROM stats
WHERE timestamp >= $1
  AND timestamp < $2
ORDER BY timestamp, node`[1:],
	}, {
		should: "generate a more complex sampler query for SQLite",
		givenConstraints: []stats.Constraint{
			{Key: "node", Operator: stats.EQ, Value: "foo"},
			{Key: "request.id", Operator: stats.IN, Value: []string{"1", "2"}},
			{Key: "timestamp", Operator: stats.GTE, Value: time.Now()},
			{Key: "timestamp", Operator: stats.LT, Value: time.Now()},
		},
		givenVars: []string{
			"node", "timestamp", "request.size", "request.id",
		},
		givenDriver: s.sqlite,
		expect: `
SELECT
  node
  , timestamp
  , request_size
  , request_id
FROM stats
WHERE node = ?
  AND request_id IN ?
  AND timestamp >= ?
  AND timestamp < ?
ORDER BY timestamp, node`[1:],
	}, {
		should: "generate a more complex sampler query for Postgres",
		givenConstraints: []stats.Constraint{
			{Key: "node", Operator: stats.EQ, Value: "foo"},
			{Key: "request.id", Operator: stats.IN, Value: []string{"1", "2"}},
			{Key: "timestamp", Operator: stats.GTE, Value: time.Now()},
			{Key: "timestamp", Operator: stats.LT, Value: time.Now()},
		},
		givenVars: []string{
			"node", "timestamp", "request.size", "request.id",
		},
		givenDriver: s.postgres,
		expect: `
SELECT
  node
  , timestamp
  , request_size
  , request_id
FROM stats
WHERE node = $1
  AND request_id IN $2
  AND timestamp >= $3
  AND timestamp < $4
ORDER BY timestamp, node`[1:],
	}} {
		c.Logf("test %d: should %s", i, test.should)
		sq := &sql.SQL{DB: test.givenDriver}

		got := sql.SampleQuery(
			sq.Parameters,
			test.givenConstraints,
			test.givenVars,
		)

		c.Check(got, gc.Equals, test.expect)
	}
}

func (s *SQLSuite) TestSample(c *gc.C) {
	tNow := time.Now()

	for i, test := range []struct {
		should            string
		driver            *sqlx.DB
		given             map[string][]stats.Point
		givenConstraints  []stats.Constraint
		givenMeasurements []string
		expect            stats.Result
		expectError       string
	}{{
		should: "fail with no measurements",
		driver: s.sqlite,
		given: map[string][]stats.Point{
			"global": {samplePoint("simple", tNow)},
		},
		expectError: "no vars given",
	}, {
		should: "fail with unknown var ",
		driver: s.sqlite,
		given: map[string][]stats.Point{
			"global": {samplePoint("simple", tNow)},
		},
		givenMeasurements: []string{"SELECT"},
		expectError:       `unknown var "SELECT"`,
	}, {
		should: "work with nil Constraints",
		driver: s.sqlite,
		given: map[string][]stats.Point{
			"global": {samplePoint("simple", tNow)},
		},
		givenMeasurements: []string{"node"},
		expect:            stats.Result{{Node: "global"}},
	}, {
		should: "group and order correctly",
		driver: s.postgres,
		given: map[string][]stats.Point{
			"global": {
				samplePoint("simple", tNow.Add(-1*time.Second).UTC()),
				samplePoint("simple", tNow.UTC()),
				samplePoint("simple", tNow.Add(1*time.Second).UTC()),
			},
			"node1": {samplePoint("simple", tNow.UTC())},
		},
		givenConstraints: []stats.Constraint{
			{Key: "timestamp", Operator: stats.GTE, Value: tNow.Add(-2 * time.Second).UTC()},
			{Key: "timestamp", Operator: stats.LT, Value: tNow.Add(2 * time.Second).UTC()},
		},
		givenMeasurements: []string{"node", "timestamp"},
		expect: stats.Result{
			{Node: "global", Timestamp: tNow.Add(-1 * time.Second).UTC()},
			{Node: "global", Timestamp: tNow.UTC()},
			{Node: "node1", Timestamp: tNow.UTC()},
			{Node: "global", Timestamp: tNow.Add(1 * time.Second).UTC()},
		},
	}, {
		should: "group, order, and sample values correctly",
		driver: s.postgres,
		given: map[string][]stats.Point{
			"global": {
				samplePoint("simple", tNow.Add(-3*time.Second)),
				samplePoint("simple", tNow.Add(-1*time.Second)),
				samplePoint("simple", tNow),
				samplePoint("simple", tNow.Add(1*time.Second)),
			},
			"node1": {
				samplePoint("simple", tNow.Add(-500*time.Millisecond)),
				samplePoint("simple", tNow.Add(-250*time.Millisecond)),
				samplePoint("simple", tNow),
			},
		},
		givenConstraints: []stats.Constraint{
			{Key: "timestamp", Operator: stats.GTE, Value: tNow.Add(-2 * time.Second)},
			{Key: "timestamp", Operator: stats.LT, Value: tNow.Add(2 * time.Second)},
		},
		givenMeasurements: []string{"node", "timestamp", "request.size"},
		expect: stats.Result{{
			Node:      "global",
			Timestamp: tNow.Add(-1 * time.Second).UTC(),
			Values: mapOnly(
				samplePoint("simple", tNow).Values,
				"request.size",
			),
		}, {
			Node:      "node1",
			Timestamp: tNow.Add(-500 * time.Millisecond).UTC(),
			Values: mapOnly(
				samplePoint("simple", tNow).Values,
				"request.size",
			),
		}, {
			Node:      "node1",
			Timestamp: tNow.Add(-250 * time.Millisecond).UTC(),
			Values: mapOnly(
				samplePoint("simple", tNow).Values,
				"request.size",
			),
		}, {
			Node:      "global",
			Timestamp: tNow.UTC(),
			Values: mapOnly(
				samplePoint("simple", tNow).Values,
				"request.size",
			),
		}, {
			Node:      "node1",
			Timestamp: tNow.UTC(),
			Values: mapOnly(
				samplePoint("simple", tNow).Values,
				"request.size",
			),
		}, {
			Node:      "global",
			Timestamp: tNow.Add(1 * time.Second).UTC(),
			Values: mapOnly(
				samplePoint("simple", tNow).Values,
				"request.size",
			),
		}},
	}, {
		should: "restrict results correctly using stats.Constraints",
		driver: s.postgres,
		given: map[string][]stats.Point{
			"global": {
				samplePoint("simple", tNow.Add(-1*time.Second)),
				samplePoint("simple", tNow),
				samplePoint("simple", tNow.Add(1*time.Second)),
			},
			"node1": {
				samplePoint("simple", tNow.Add(-500*time.Millisecond)),
				samplePoint("simple", tNow.Add(-250*time.Millisecond)),
				samplePoint("simple", tNow),
			},
		},
		givenConstraints: []stats.Constraint{
			{Key: "timestamp", Operator: stats.GTE, Value: tNow.Add(-2 * time.Second)},
			{Key: "timestamp", Operator: stats.LT, Value: tNow.Add(2 * time.Second)},
			{Key: "node", Operator: stats.EQ, Value: "global"},
		},
		givenMeasurements: []string{"timestamp", "response.time"},
		expect: stats.Result{{
			Timestamp: tNow.Add(-1 * time.Second).UTC(),
			Values: mapOnly(
				samplePoint("simple", tNow).Values,
				"response.time",
			),
		}, {
			Timestamp: tNow.UTC(),
			Values: mapOnly(
				samplePoint("simple", tNow).Values,
				"response.time",
			),
		}, {
			Timestamp: tNow.Add(1 * time.Second).UTC(),
			Values: mapOnly(
				samplePoint("simple", tNow).Values,
				"response.time",
			),
		}},
	}} {
		c.Logf("test %d: should %s", i, test.should)
		sq := &sql.SQL{DB: test.driver}

		_, er := sq.Exec(`DELETE FROM stats`)
		c.Assert(er, gc.IsNil)

		oldNAME := sq.NAME

		for name, points := range test.given {
			sq.NAME = name
			c.Assert(sq.Log(points...), gc.IsNil)
		}
		sq.NAME = oldNAME

		got, err := sq.Sample(
			test.givenConstraints,
			test.givenMeasurements...,
		)

		if test.expectError != "" {
			c.Check(err, gc.ErrorMatches, test.expectError)
			continue
		}

		c.Assert(err, gc.IsNil)

		c.Assert(len(got), gc.Equals, len(test.expect))

		for index, pointBack := range got {
			pointExpected := test.expect[index]
			c.Assert(pointBack.Node, gc.Equals, pointExpected.Node)
			c.Assert(pointBack.Err, gc.Equals, pointExpected.Err)
			c.Assert(pointBack.Timestamp.Format(time.RFC3339), gc.Equals, pointExpected.Timestamp.Format(time.RFC3339))
			c.Assert(pointBack.Values, gc.DeepEquals, pointExpected.Values)
		}
	}
}
