package admin_test

import (
	"gateway/errors"
	"gateway/model"
	"gateway/model/testing"
	"gateway/stats"

	gc "gopkg.in/check.v1"
)

func (a *AdminSuite) TestSampleBeforeValidate(c *gc.C) {
	testing.PrepareAccount(c, a.db, testing.JeffAccount)

	for i, t := range []struct {
		should   string
		given    model.Sample
		isInsert bool
		expect   errors.Errors
	}{{
		should: "give error for invalid Operator",
		given: model.Sample{
			Constraints: []stats.Constraint{{Operator: "foo"}},
		},
		isInsert: true,
		expect:   errors.Errors{"operator": {`"foo" is not a valid operator`}},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		c.Check(
			t.given.BeforeValidate(t.isInsert),
			gc.DeepEquals,
			t.expect,
		)
	}
}
