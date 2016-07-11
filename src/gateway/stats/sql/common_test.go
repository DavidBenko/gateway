package sql_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"gateway/stats"
	"gateway/stats/sql"

	"github.com/jmoiron/sqlx"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gc.TestingT(t) }

type SQLSuite struct {
	sqlite   *sqlx.DB
	postgres *sqlx.DB
}

const (
	defaultPgString = "gateway_test"
)

var (
	_ = gc.Suite(&SQLSuite{})

	pgDBName = func() string {
		if s := os.Getenv("POSTGRES_DB_NAME"); s != "" {
			return s
		}
		return defaultPgString
	}()

	pgConnString = strings.Join([]string{
		"dbname=" + pgDBName,
		"sslmode=disable",
	}, " ")

	_ = stats.Logger(&sql.SQL{})
	_ = stats.Sampler(&sql.SQL{})
)

func (s *SQLSuite) TearDownTest(c *gc.C) {
	s.teardown(c)
}

func (s *SQLSuite) SetUpTest(c *gc.C) {
	s.setup(c)
}

// mapOnly returns a copy of the given map containing only the given keys.
func mapOnly(m1 map[string]interface{}, ks ...string) map[string]interface{} {
	res := make(map[string]interface{})
	for _, k := range ks {
		res[k] = m1[k]
	}
	return res
}

// mapWithout returns a copy of the given map without the given keys.
func mapWithout(m1 map[string]interface{}, ks ...string) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range m1 {
		res[k] = v
	}
	for _, k := range ks {
		delete(res, k)
	}
	return res
}

func samplePoint(name string, tst time.Time) stats.Point {
	return map[string]stats.Point{
		"simple": stats.Point{
			Timestamp: tst,
			Values: map[string]interface{}{
				"request.size":                  0,
				"request.id":                    "1234",
				"api.id":                        int64(1),
				"api.name":                      "text",
				"response.time":                 50,
				"response.size":                 500,
				"response.status":               http.StatusOK,
				"response.error":                "",
				"host.id":                       int64(2),
				"host.name":                     "text",
				"proxy.id":                      int64(2),
				"proxy.name":                    "text",
				"proxy.env.id":                  int64(2),
				"proxy.env.name":                "text",
				"proxy.route.path":              "text",
				"proxy.route.verb":              "text",
				"proxy.group.id":                int64(2),
				"proxy.group.name":              "text",
				"remote_endpoint.response.time": 2,
			},
		},
		"simple1": stats.Point{
			Timestamp: tst,
			Values: map[string]interface{}{
				"request.size":                  0,
				"request.id":                    "1234",
				"api.id":                        int64(1),
				"api.name":                      "text",
				"response.time":                 50,
				"response.size":                 500,
				"response.status":               http.StatusOK,
				"response.error":                "",
				"host.id":                       int64(2),
				"host.name":                     "text",
				"proxy.id":                      int64(2),
				"proxy.name":                    "text",
				"proxy.env.id":                  int64(2),
				"proxy.env.name":                "text",
				"proxy.route.path":              "text",
				"proxy.route.verb":              "text",
				"proxy.group.id":                int64(2),
				"proxy.group.name":              "text",
				"remote_endpoint.response.time": 2,
			},
		},
		"simple2": stats.Point{
			Timestamp: tst,
			Values: map[string]interface{}{
				"request.size":                  10,
				"request.id":                    "1234",
				"api.id":                        int64(1),
				"api.name":                      "text",
				"response.time":                 60,
				"response.size":                 500,
				"response.status":               http.StatusOK,
				"response.error":                "",
				"host.id":                       int64(2),
				"host.name":                     "text",
				"proxy.id":                      int64(2),
				"proxy.name":                    "text",
				"proxy.env.id":                  int64(2),
				"proxy.env.name":                "text",
				"proxy.route.path":              "text",
				"proxy.route.verb":              "text",
				"proxy.group.id":                int64(2),
				"proxy.group.name":              "text",
				"remote_endpoint.response.time": 2,
			},
		},
		"simple3": stats.Point{
			Timestamp: tst,
			Values: map[string]interface{}{
				"request.size":                  20,
				"request.id":                    "1234",
				"api.id":                        int64(1),
				"api.name":                      "text",
				"response.time":                 70,
				"response.size":                 500,
				"response.status":               http.StatusOK,
				"response.error":                "",
				"host.id":                       2,
				"host.name":                     "text",
				"proxy.id":                      2,
				"proxy.name":                    "text",
				"proxy.env.id":                  2,
				"proxy.env.name":                "text",
				"proxy.route.path":              "text",
				"proxy.route.verb":              "text",
				"proxy.group.id":                2,
				"proxy.group.name":              "text",
				"remote_endpoint.response.time": 2,
			},
		},
	}[name]
}

func sampleRow(name, node string, tst time.Time) stats.Row {
	return map[string]stats.Row{
		"simple": {
			Node:      node,
			Timestamp: tst,
			Values: map[string]interface{}{
				"request.size":                  0,
				"request.id":                    "1234",
				"api.id":                        int64(1),
				"api.name":                      "text",
				"response.time":                 50,
				"response.size":                 500,
				"response.status":               http.StatusOK,
				"response.error":                "",
				"host.id":                       int64(2),
				"host.name":                     "text",
				"proxy.id":                      int64(2),
				"proxy.name":                    "text",
				"proxy.env.id":                  int64(2),
				"proxy.env.name":                "text",
				"proxy.route.path":              "text",
				"proxy.route.verb":              "text",
				"proxy.group.id":                int64(2),
				"proxy.group.name":              "text",
				"remote_endpoint.response.time": 2,
			},
		},
		"simple1": {
			Node:      node,
			Timestamp: tst,
			Values: map[string]interface{}{
				"request.size":                  0,
				"request.id":                    "1234",
				"api.id":                        int64(1),
				"api.name":                      "text",
				"response.time":                 50,
				"response.size":                 500,
				"response.status":               http.StatusOK,
				"response.error":                "",
				"host.id":                       int64(2),
				"host.name":                     "text",
				"proxy.id":                      int64(2),
				"proxy.name":                    "text",
				"proxy.env.id":                  int64(2),
				"proxy.env.name":                "text",
				"proxy.route.path":              "text",
				"proxy.route.verb":              "text",
				"proxy.group.id":                int64(2),
				"proxy.group.name":              "text",
				"remote_endpoint.response.time": 2,
			},
		},
		"simple2": {
			Node:      node,
			Timestamp: tst,
			Values: map[string]interface{}{
				"request.size":                  10,
				"request.id":                    "1234",
				"api.id":                        int64(1),
				"api.name":                      "text",
				"response.time":                 60,
				"response.size":                 500,
				"response.status":               http.StatusOK,
				"response.error":                "",
				"host.id":                       int64(2),
				"host.name":                     "text",
				"proxy.id":                      int64(2),
				"proxy.name":                    "text",
				"proxy.env.id":                  int64(2),
				"proxy.env.name":                "text",
				"proxy.route.path":              "text",
				"proxy.route.verb":              "text",
				"proxy.group.id":                int64(2),
				"proxy.group.name":              "text",
				"remote_endpoint.response.time": 2,
			},
		},
		"simple3": {
			Node:      node,
			Timestamp: tst,
			Values: map[string]interface{}{
				"request.size":                  20,
				"request.id":                    "1234",
				"api.id":                        int64(1),
				"api.name":                      "text",
				"response.time":                 70,
				"response.size":                 500,
				"response.status":               http.StatusOK,
				"response.error":                "",
				"host.id":                       int64(2),
				"host.name":                     "text",
				"proxy.id":                      int64(2),
				"proxy.name":                    "text",
				"proxy.env.id":                  int64(2),
				"proxy.env.name":                "text",
				"proxy.route.path":              "text",
				"proxy.route.verb":              "text",
				"proxy.group.id":                int64(2),
				"proxy.group.name":              "text",
				"remote_endpoint.response.time": 2,
			},
		},
	}[name]
}

func (s *SQLSuite) setup(c *gc.C) {
	c.Log("    >>DB: Connecting to in-memory sqlite3 database")

	sqliteDB, err := sqlx.Open("sqlite3", ":memory:")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sqliteDB.Ping(), jc.ErrorIsNil)
	c.Assert(sql.Migrate(sqliteDB, sql.SQLite3), jc.ErrorIsNil)

	c.Logf("    >>DB: Connecting to pq using connection string %q", pgConnString)

	pgDB, err := sqlx.Open("postgres", pgConnString)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(pgDB.Ping(), jc.ErrorIsNil)
	c.Assert(sql.Migrate(pgDB, sql.Postgres), jc.ErrorIsNil)

	s.sqlite, s.postgres = sqliteDB, pgDB
}

func (s *SQLSuite) teardown(c *gc.C) {
	c.Log("    >>DB: Cleaning up Postgres database")
	for _, table := range []string{
		"stats",
		"stats_schema",
	} {
		c.Logf("    >>DB: dropping Postgres table %q", table)
		_, err := s.postgres.Exec(fmt.Sprintf(
			"DROP TABLE IF EXISTS %s;", table,
		))

		c.Assert(err, jc.ErrorIsNil)

		c.Logf("    >>DB: dropping SQLite3 table %q", table)
		_, err = s.sqlite.Exec(fmt.Sprintf(
			"DROP TABLE IF EXISTS %s;", table,
		))

		c.Assert(err, jc.ErrorIsNil)
	}

	c.Log("    >>DB: Closing sqlite3 connection")
	c.Assert(s.sqlite.Close(), jc.ErrorIsNil)

	c.Log("    >>DB: Closing postgres connection")
	c.Assert(s.postgres.Close(), jc.ErrorIsNil)
}
