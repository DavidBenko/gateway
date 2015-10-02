package sql_test

import (
	"fmt"
	"gateway/db"
	"testing"

	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type SQLSuite struct{}

var _ = gc.Suite(&SQLSuite{})

func (s *SQLSuite) TestNeedsUpdate(c *gc.C) {
	self := pqSpecs()["simple"]
	for i, t := range []struct {
		should      string
		given       db.Specifier
		compare     db.Specifier
		expect      bool
		expectPanic string
	}{{
		should:  "not error on a self-check",
		given:   self,
		compare: self,
	}, {
		should:      "not work comparing different types",
		given:       pqSpecs()["simple"],
		compare:     mysqlSpecs()["simple"],
		expectPanic: "tried to compare wrong database kinds: *sql.MySQLSpec and *sql.PostgresSpec",
	}, {
		should:      "fail to compare nil specs",
		given:       pqSpecs()["simple"],
		expectPanic: "tried to compare to nil db.Specifier!",
	}} {
		msg := fmt.Sprintf("Test %d: should %s", i, t.should)
		if t.expectPanic != "" {
			msg += " (expect panic)"
		}

		c.Logf(msg)

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

			c1, c2 := t.given, t.compare
			switch {
			case c1 == nil && c2 == nil:
				c.Log("tried to compare a nil spec to a nil spec")
				c.FailNow()
			case c1 == nil:
				result := c2.NeedsUpdate(c1)
				c.Check(result, gc.Equals, t.expect)
			case c2 == nil:
				result := c1.NeedsUpdate(c2)
				c.Check(result, gc.Equals, t.expect)
			default:
				result := c2.NeedsUpdate(c1)
				c.Check(result, gc.Equals, t.expect)
				result = c1.NeedsUpdate(c2)
				c.Check(result, gc.Equals, t.expect)
			}
		}()
	}
}
