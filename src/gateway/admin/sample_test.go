package admin_test

import (
	"gateway/admin"
	"gateway/model"
	"gateway/model/testing"
	"gateway/stats"

	gc "gopkg.in/check.v1"
)

func (a *AdminSuite) TestSampleBeforeValidate(c *gc.C) {
	acc1 := testing.PrepareAccount(c, a.db, testing.JeffAccount)
	user1 := testing.PrepareUser(c, a.db, acc1.ID, testing.JeffUser)
	testing.PrepareAPI(c, a.db, acc1.ID, user1.ID, testing.API2)
	controller := &admin.SamplesController{}

	for i, t := range []struct {
		should      string
		given       *model.Sample
		expect      string
		expectError string
	}{{
		should: "give error for invalid Operator",
		given: &model.Sample{
			Name:      "Sample1",
			AccountID: acc1.ID,
			UserID:    user1.ID,
			Constraints: []stats.Constraint{
				{Key: "api.id", Operator: "foo", Value: int64(5)},
			},
		},
		expectError: `invalid operator "foo" for single api.id value, use "EQ"`,
	}} {
		c.Logf("test %d: should %s", i, t.should)

		tx, err := a.db.Begin()
		c.Assert(err, gc.IsNil)
		gotten := controller.BeforeValidate(t.given, tx)
		if t.expectError != "" {
			c.Check(gotten, gc.ErrorMatches, t.expectError)
		} else {
			c.Check(gotten, gc.DeepEquals, t.expect)
		}
		tx.Commit()

	}
}
