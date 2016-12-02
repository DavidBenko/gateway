package sql_test

import (
	"gateway/stats/sql"

	gc "gopkg.in/check.v1"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func (s *SQLSuite) TestParameters(c *gc.C) {
	for i, test := range []struct {
		driver *sqlx.DB
		given  int
		expect []string
	}{
		{s.sqlite, -1, nil},
		{s.sqlite, 0, nil},
		{s.sqlite, 1, []string{"?"}},
		{s.sqlite, 2, []string{"?", "?"}},
		{s.sqlite, 11, []string{
			"?", "?", "?", "?", "?",
			"?", "?", "?", "?", "?", "?",
		}},
		{s.postgres, -1, nil},
		{s.postgres, 0, nil},
		{s.postgres, 1, []string{"$1"}},
		{s.postgres, 2, []string{"$1", "$2"}},
		{s.postgres, 11, []string{
			"$1", "$2", "$3", "$4", "$5",
			"$6", "$7", "$8", "$9", "$10", "$11",
		}},
	} {
		c.Logf("test %d:\n  given %d, expect %v", i,
			test.given, test.expect,
		)
		got := (&sql.SQL{DB: test.driver}).Parameters(test.given)
		c.Check(got, gc.DeepEquals, test.expect)
	}
}
