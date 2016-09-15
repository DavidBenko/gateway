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
	given := &model.Sample{
		Name:        "Sample1",
		Constraints: []stats.Constraint{{Key: "api.id", Operator: "foo", Value: int64(5)}},
		AccountID:   acc1.ID,
		UserID:      user1.ID,
	}
	expected := `invalid operator "foo" for single api.id value, use "EQ"`
	c.Logf("test 0: should %s", "give error for invalid Operator")
	tx, err := a.db.Begin()
	c.Assert(err, gc.IsNil)
	gotten := controller.BeforeValidate(given, tx)
	c.Check(gotten, gc.ErrorMatches, expected)
	tx.Commit()
}
