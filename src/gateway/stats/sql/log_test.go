package sql_test

import (
	"errors"
	"fmt"
	"time"

	gwerr "gateway/errors"
	"gateway/stats"
	"gateway/stats/sql"

	"github.com/jmoiron/sqlx"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func (s *SQLSuite) TestLogQuery(c *gc.C) {
	for i, t := range []struct {
		should      string
		given       int
		givenNode   string
		givenDriver *sqlx.DB
		expect      string
	}{{
		should:      "generate a correct query for SQLite",
		given:       1,
		givenNode:   "global",
		givenDriver: s.sqlite,
		expect: `
INSERT INTO stats (
  node
  , timestamp
  , ms
  , request_size
  , request_id
  , response_time
  , response_size
  , response_status
  , response_error
) VALUES
  ('global', ?, ?, ?, ?, ?, ?, ?, ?)
`[1:],
	}, {
		should:      "generate a correct query for multi-point SQLite",
		given:       3,
		givenNode:   "global",
		givenDriver: s.sqlite,
		expect: `
INSERT INTO stats (
  node
  , timestamp
  , ms
  , request_size
  , request_id
  , response_time
  , response_size
  , response_status
  , response_error
) VALUES
  ('global', ?, ?, ?, ?, ?, ?, ?, ?)
  , ('global', ?, ?, ?, ?, ?, ?, ?, ?)
  , ('global', ?, ?, ?, ?, ?, ?, ?, ?)
`[1:],
	}, {
		should:      "generate a correct query for Postgres",
		given:       1,
		givenNode:   "global",
		givenDriver: s.postgres,
		expect: `
INSERT INTO stats (
  node
  , timestamp
  , ms
  , request_size
  , request_id
  , response_time
  , response_size
  , response_status
  , response_error
) VALUES
  ('global', $1, $2, $3, $4, $5, $6, $7, $8)
`[1:],
	}, {
		should:      "generate a correct query for multi-point Postgres",
		given:       3,
		givenNode:   "global",
		givenDriver: s.postgres,
		expect: `
INSERT INTO stats (
  node
  , timestamp
  , ms
  , request_size
  , request_id
  , response_time
  , response_size
  , response_status
  , response_error
) VALUES
  ('global', $1, $2, $3, $4, $5, $6, $7, $8)
  , ('global', $9, $10, $11, $12, $13, $14, $15, $16)
  , ('global', $17, $18, $19, $20, $21, $22, $23, $24)
`[1:],
	}} {
		c.Logf("test %d: should %s", i, t.should)
		sq := &sql.SQL{ID: t.givenNode, DB: t.givenDriver}

		got := sql.LogQuery(
			sq.Parameters,
			t.givenNode,
			t.given,
		)

		c.Check(got, gc.Equals, t.expect)
	}
}

func (s *SQLSuite) TestGetArgs(c *gc.C) {
	tNow := time.Now()

	for i, t := range []struct {
		should    string
		given     []stats.Point
		expect    []interface{}
		expectErr string
	}{{
		should:    "return error for a nil slice",
		expectErr: `must pass at least one stats.Point`,
	}, {
		should: "return error for a Point missing measurements",
		given: []stats.Point{{
			Timestamp: tNow,
			Values:    map[string]interface{}{"request_time": 0},
		}},
		expectErr: `point missing measurement "request.size"`,
	}, {
		should: "get args for stats.Point slice of 1 element",
		given:  []stats.Point{samplePoint("simple", tNow)},
		expect: []interface{}{
			tNow.UTC(), sql.DayMillis(tNow.UTC()),
			0, "1234", 50, 500, 200, "",
		},
	}, {
		should: "get args for stats.Point slice of several elements",
		given: []stats.Point{
			samplePoint("simple1", tNow),
			samplePoint("simple2", tNow),
		},
		expect: []interface{}{
			tNow.UTC(), sql.DayMillis(tNow.UTC()),
			0, "1234", 50, 500, 200, "",
			tNow.UTC(), sql.DayMillis(tNow.UTC()),
			10, "1234", 60, 500, 200, "",
		},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		got, err := sql.GetArgs(t.given...)
		if t.expectErr != "" {
			c.Check(err, gc.ErrorMatches, t.expectErr)
			continue
		}

		c.Assert(err, jc.ErrorIsNil)

		// Should not have mutated the given slice.
		c.Assert(t.given, jc.DeepEquals, t.given)

		c.Check(got, jc.DeepEquals, t.expect)
	}
}

func (s *SQLSuite) TestLog(c *gc.C) {
	tNow := time.Now()

	for i, t := range []struct {
		should      string
		node        string
		timestamp   time.Time
		points      []stats.Point
		expect      stats.Result
		expectError string
	}{{
		should: "break if unknown measurement",
		points: []stats.Point{{
			Timestamp: tNow,
			Values:    map[string]interface{}{"something": 0},
		}},
		expectError: `failed to log: failed to get args for stats ` +
			`query: point missing measurement "request.size"`,
	}, {
		should:    "log a single point",
		timestamp: tNow,
		points:    []stats.Point{samplePoint("simple", tNow)},
		expect:    []stats.Row{sampleRow("simple1", "global", tNow.UTC())},
	}, {
		should:    "log multiple points",
		timestamp: tNow,
		points: []stats.Point{
			samplePoint("simple1", tNow),
			samplePoint("simple2", tNow.Add(1*time.Second)),
			samplePoint("simple3", tNow.Add(2*time.Second)),
		},
		expect: []stats.Row{
			sampleRow("simple1", "global", tNow.UTC()),
			sampleRow("simple2", "global", tNow.Add(1*time.Second).UTC()),
			sampleRow("simple3", "global", tNow.Add(2*time.Second).UTC()),
		},
	}} {
		c.Logf("test %d: should %s", i, t.should)
		s.teardown(c)
		s.setup(c)

		for _, db := range []*sqlx.DB{
			s.sqlite, s.postgres,
		} {
			c.Logf("  testing with driver %q", db.DriverName())
			sq := &sql.SQL{ID: t.node, DB: db}

			result, err := testLog(
				sq,
				t.timestamp,
				t.points...,
			)

			if t.expectError != "" {
				c.Check(err, gc.ErrorMatches, t.expectError)
				continue
			}

			c.Assert(err, jc.ErrorIsNil)

			expect := make(stats.Result, len(t.expect))
			for i, r := range t.expect {
				fixed := r
				// Note that nanosecond timestamp precision is
				// not completely exact.  Reduce precision by
				// rounding up and truncating last 4 digits.
				fns := fixed.Timestamp.Nanosecond()
				fns = int((fns+500)/1000) * 1000
				fixed.Timestamp = time.Date(
					fixed.Timestamp.Year(),
					fixed.Timestamp.Month(),
					fixed.Timestamp.Day(),
					fixed.Timestamp.Hour(),
					fixed.Timestamp.Minute(),
					fixed.Timestamp.Second(),
					fns, // rounded nanoseconds
					fixed.Timestamp.Location(),
				)

				resRemoveNS := result[i]
				rns := resRemoveNS.Timestamp.Nanosecond()
				rns = int((rns+500)/1000) * 1000
				resRemoveNS.Timestamp = time.Date(
					resRemoveNS.Timestamp.Year(),
					resRemoveNS.Timestamp.Month(),
					resRemoveNS.Timestamp.Day(),
					resRemoveNS.Timestamp.Hour(),
					resRemoveNS.Timestamp.Minute(),
					resRemoveNS.Timestamp.Second(),
					rns, // rounded nanoseconds
					resRemoveNS.Timestamp.Location(),
				)
				result[i] = resRemoveNS
				expect[i] = fixed
			}

			c.Check(result, jc.DeepEquals, expect)
		}
	}
}

func testLog(
	s *sql.SQL,
	timestamp time.Time,
	points ...stats.Point,
) (stats.Result, error) {
	if err := s.Log(points...); err != nil {
		return nil, gwerr.NewWrapped("failed to log", err)
	}

	ID := "global"
	if s.ID != "" {
		ID = s.ID
	}

	rows, err := s.Queryx(fmt.Sprintf(`
SELECT
  node
  , timestamp
  , request_size
  , request_id
  , response_time
  , response_size
  , response_status
  , response_error
FROM stats
WHERE node = '%s'`[1:], ID))

	switch {
	case err != nil:
		return nil, gwerr.NewWrapped("failed to select", err)
	case rows == nil:
		return nil, errors.New("no rows for stats query")
	}

	defer rows.Close()

	var result stats.Result

	for rowNum := 0; rows.Next(); rowNum++ {
		var row sql.Row

		if err = rows.StructScan(&row); err != nil {
			return nil, gwerr.NewWrapped("failed to scan", err)
		}

		result = append(result, stats.Row{
			Node:      row.Node,
			Timestamp: row.Timestamp.UTC(),
			Values: map[string]interface{}{
				"request.size":    row.RequestSize,
				"request.id":      row.RequestID,
				"response.time":   row.ResponseTime,
				"response.size":   row.ResponseSize,
				"response.status": row.ResponseStatus,
				"response.error":  row.ResponseError,
			},
		})
	}

	if err = rows.Err(); err != nil {
		return nil, gwerr.NewWrapped("rows had error", err)
	}

	return result, nil
}
