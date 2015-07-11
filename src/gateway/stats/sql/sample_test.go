package sql_test

import (
	"fmt"
	"time"

	"gateway/stats"
	"gateway/stats/sql"

	"github.com/davecgh/go-spew/spew"
	"github.com/jmoiron/sqlx"
	gc "gopkg.in/check.v1"
)

func (s *SQLSuite) TestSampleQuery(c *gc.C) {
	for i, test := range []struct {
		should      string
		givenVars   []string
		givenTags   map[string]interface{}
		givenFrom   time.Time
		givenTo     time.Time
		givenDriver *sqlx.DB
		expect      string
	}{{
		should:      "generate a simple sampler query for SQLite",
		givenVars:   []string{"request.size", "request.id"},
		givenDriver: s.sqlite,
		expect: `
SELECT
  request_size
  , request_id
FROM stats
WHERE timestamp >= ? AND timestamp < ?
ORDER BY timestamp, node`[1:],
	}, {
		should:      "generate a simple sampler query for Postgres",
		givenVars:   []string{"request.size", "request.id"},
		givenDriver: s.postgres,
		expect: `
SELECT
  request_size
  , request_id
FROM stats
WHERE timestamp >= $1 AND timestamp < $2
ORDER BY timestamp, node`[1:],
	}, {
		should: "generate a more complex sampler query for SQLite",
		givenTags: map[string]interface{}{
			"node":       "global",
			"request.id": "1234",
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
  AND request_id = ?
  AND timestamp >= ? AND timestamp < ?
ORDER BY timestamp, node`[1:],
	}} {
		c.Logf("test %d: should %s", i, test.should)
		sq := &sql.SQL{DB: test.givenDriver}

		var tagParams []string
		for t := range test.givenTags {
			tagParams = append(tagParams, t)
		}

		got := sql.SampleQuery(
			sq.Parameters,
			test.givenVars,
			tagParams,
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
		givenTags         map[string]interface{}
		from              time.Time
		to                time.Time
		givenMeasurements []string
		expect            stats.Result
		expectError       string
	}{{
		should:    "fail with bad time constraints",
		driver:    s.sqlite,
		givenTags: map[string]interface{}{"node": "global"},
		given: map[string][]stats.Point{
			"global": {samplePoint("simple", tNow)},
		},
		from: tNow.Add(time.Second),
		to:   tNow.Add(-1 * time.Second),
		expectError: fmt.Sprintf(
			"time %s is not after %s",
			tNow.Add(-1*time.Second),
			tNow.Add(time.Second),
		),
	}, {
		should:    "fail with no measurements",
		driver:    s.sqlite,
		givenTags: map[string]interface{}{"node": "global"},
		given: map[string][]stats.Point{
			"global": {samplePoint("simple", tNow)},
		},
		from:        tNow.Add(-1 * time.Second),
		to:          tNow.Add(time.Second),
		expectError: "no measurements given",
	}, {
		should: "work with nil tags",
		driver: s.sqlite,
		given: map[string][]stats.Point{
			"global": {samplePoint("simple", tNow)},
		},
		from:              tNow.Add(-1 * time.Second),
		to:                tNow.Add(time.Second),
		givenMeasurements: []string{"node"},
		expect:            stats.Result{{Node: "global"}},
	}, {
		should: "group and order correctly",
		driver: s.sqlite,
		given: map[string][]stats.Point{
			"global": {
				samplePoint("simple", tNow.Add(-1*time.Second)),
				samplePoint("simple", tNow),
				samplePoint("simple", tNow.Add(1*time.Second)),
			},
			"node1": {samplePoint("simple", tNow)},
		},
		from:              tNow.Add(-2 * time.Second),
		to:                tNow.Add(2 * time.Second),
		givenMeasurements: []string{"node", "timestamp"},
		expect: stats.Result{
			{Node: "global", Timestamp: tNow.Add(-1 * time.Second).UTC()},
			{Node: "global", Timestamp: tNow.UTC()},
			{Node: "node1", Timestamp: tNow.UTC()},
			{Node: "global", Timestamp: tNow.Add(1 * time.Second).UTC()},
		},
	}, {
		should: "group, order, and sample values correctly",
		driver: s.sqlite,
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
		from:              tNow.Add(-2 * time.Second),
		to:                tNow.Add(2 * time.Second),
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
		should: "restrict results on tags correctly",
		driver: s.sqlite,
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
		from: tNow.Add(-2 * time.Second),
		to:   tNow.Add(2 * time.Second),
		givenTags: map[string]interface{}{
			"node": "global",
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

		_, err := sq.Exec(`DELETE FROM stats`)
		c.Assert(err, gc.IsNil)

		oldID := sq.ID
		for node, points := range test.given {
			sq.ID = node
			c.Check(sq.Log(points...), gc.IsNil)
			if c.Failed() {
				continue
			}
		}
		sq.ID = oldID

		got, err := sq.Sample(
			test.givenTags,
			test.from,
			test.to,
			test.givenMeasurements...,
		)

		if test.expectError != "" {
			c.Check(err, gc.ErrorMatches, test.expectError)
			continue
		}

		c.Assert(err, gc.IsNil)

		c.Check(got, gc.DeepEquals, test.expect)
		if c.Failed() {
			c.Logf("got: %s\nexpected: %s",
				spew.Sdump(got),
				spew.Sdump(test.expect),
			)
		}
	}
}
