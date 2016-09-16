package admin_test

import (
	"gateway/admin"
	"gateway/model"
	"gateway/model/testing"
	"gateway/stats"
	"regexp"

	gc "gopkg.in/check.v1"
)

func (a *AdminSuite) TestSampleBeforeValidate(c *gc.C) {
	acc1 := testing.PrepareAccount(c, a.db, testing.JeffAccount)
	user1 := testing.PrepareUser(c, a.db, acc1.ID, testing.JeffUser)
	api1 := testing.PrepareAPI(c, a.db, acc1.ID, user1.ID, testing.API2)
	controller := &admin.SamplesController{}
	for i, t := range []struct {
		should      string
		given       *model.Sample
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
	}, {
		should: "give error for invalid value",
		given: &model.Sample{
			Name:      "Sample1",
			AccountID: acc1.ID,
			UserID:    user1.ID,
			Constraints: []stats.Constraint{
				{Key: "api.id", Operator: "EQ", Value: "5"},
			},
		},
		expectError: regexp.QuoteMeta("invalid type string for api.id value, must be int64 or []int64"),
	}, {
		should: "Work with a valid value",
		given: &model.Sample{
			Name:      "Sample1",
			AccountID: acc1.ID,
			UserID:    user1.ID,
			Constraints: []stats.Constraint{
				{Key: "api.id", Operator: stats.EQ, Value: api1.ID},
			},
		},
	},
	} {
		c.Logf("test %d: should %s", i, t.should)

		tx, err := a.db.Begin()
		c.Assert(err, gc.IsNil)
		gotten := controller.BeforeValidate(t.given, tx)
		if t.expectError != "" {
			c.Check(gotten, gc.ErrorMatches, t.expectError)
		} else {
			c.Assert(gotten, gc.IsNil)
		}
		tx.Commit()

	}
}
