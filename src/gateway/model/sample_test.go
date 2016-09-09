package model_test

import (
	"fmt"
	"gateway/errors"
	"gateway/model"
	"gateway/model/testing"
	"gateway/stats"

	gc "gopkg.in/check.v1"
)

func (m *ModelSuite) TestSampleValidateConstraints(c *gc.C) {
	acc1 := testing.PrepareAccount(c, m.db, testing.JeffAccount)
	acc2 := testing.PrepareAccount(c, m.db, testing.OtherAccount)

	user1 := testing.PrepareUser(c, m.db, acc1.ID, testing.JeffUser)
	user2 := testing.PrepareUser(c, m.db, acc2.ID, testing.OtherUser)

	api1 := testing.PrepareAPI(c, m.db, acc1.ID, user1.ID, testing.API1)
	api2 := testing.PrepareAPI(c, m.db, acc1.ID, user1.ID, testing.API2)
	api3 := testing.PrepareAPI(c, m.db, acc2.ID, user2.ID, testing.API3)

	for i, t := range []struct {
		should            string
		given             []stats.Constraint
		givenAccountID    int64
		givenUserID       int64
		isInsert          bool
		expectConstraints []stats.Constraint
		expectError       string
	}{{
		should: "give error for invalid Operator",
		given: []stats.Constraint{
			{Key: "api.id", Operator: "foo", Value: int64(5)},
		},
		isInsert: true,
		expectError: `invalid operator "foo" for single api.id value, ` +
			`use "EQ"`,
	}, {
		should: "give error for invalid Operator",
		given: []stats.Constraint{
			{Key: "api.id", Operator: "foo", Value: []int64{5}},
		},
		isInsert: true,
		expectError: `invalid operator "foo" for multiple api.id values, ` +
			`use "IN"`,
	}, {
		should: "give error for invalid accountID EQ Value type",
		given: []stats.Constraint{
			{Key: "api.id", Operator: stats.EQ, Value: "hi"},
		},
		isInsert: true,
		expectError: `invalid type string for api.id value, must be ` +
			`int64 or \[\]int64`,
	}, {
		should: "give error for accountID not owned (EQ)",
		given: []stats.Constraint{{
			Key:      "api.id",
			Operator: stats.EQ,
			Value:    int64(api3.ID),
		}},
		givenAccountID: acc1.ID,
		givenUserID:    user1.ID,
		isInsert:       true,
		expectError: fmt.Sprintf(
			`api %d not owned by user %d`, api3.ID, user1.ID,
		),
	}, {
		should: "give error for accountID not owned (IN)",
		given: []stats.Constraint{{
			Key:      "api.id",
			Operator: stats.IN,
			Value:    []int64{api1.ID, api3.ID},
		}},
		givenAccountID: acc1.ID,
		givenUserID:    user1.ID,
		isInsert:       true,
		expectError: fmt.Sprintf(
			`api %d not owned by user %d`, api3.ID, user1.ID,
		),
	}, {
		should: "give error for accountID not owned (IN)",
		given: []stats.Constraint{{
			Key:      "api.id",
			Operator: stats.IN,
			Value:    []int64{api1.ID, api3.ID},
		}},
		givenAccountID: acc1.ID,
		givenUserID:    user1.ID,
		isInsert:       true,
		expectError: fmt.Sprintf(
			`api %d not owned by user %d`, api3.ID, user1.ID,
		),
	}, {
		should: "give no error for accountID owned (EQ)",
		given: []stats.Constraint{{
			Key:      "api.id",
			Operator: stats.EQ,
			Value:    int64(api1.ID),
		}},
		givenAccountID: acc1.ID,
		givenUserID:    user1.ID,
		isInsert:       true,
		expectConstraints: []stats.Constraint{{
			Key:      "api.id",
			Operator: stats.EQ,
			Value:    int64(api1.ID),
		}},
	}, {
		should: "give no error for accountID owned (IN)",
		given: []stats.Constraint{{
			Key:      "api.id",
			Operator: stats.IN,
			Value:    []int64{api1.ID, api2.ID},
		}},
		givenAccountID: acc1.ID,
		givenUserID:    user1.ID,
		isInsert:       true,
		expectConstraints: []stats.Constraint{{
			Key:      "api.id",
			Operator: stats.IN,
			Value:    []int64{api1.ID, api2.ID},
		}},
	}, {
		should:         "inject constraint if no constraint given on api.id",
		givenAccountID: acc1.ID,
		givenUserID:    user1.ID,
		isInsert:       true,
		expectConstraints: []stats.Constraint{{
			Key:      "api.id",
			Operator: stats.IN,
			Value:    []int64{api1.ID, api2.ID},
		}},
	}, {
		should: "inject constraint if no constraint given on api.id",
		given: []stats.Constraint{{
			Key:      "request.id",
			Operator: stats.EQ,
			Value:    "fluff",
		}},
		givenAccountID: acc1.ID,
		givenUserID:    user1.ID,
		isInsert:       true,
		expectConstraints: []stats.Constraint{{
			Key:      "request.id",
			Operator: stats.EQ,
			Value:    "fluff",
		}, {
			Key:      "api.id",
			Operator: stats.IN,
			Value:    []int64{api1.ID, api2.ID},
		}},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		tx, err := m.db.Begin()
		c.Assert(err, gc.IsNil)

		given := &model.Sample{
			AccountID:   t.givenAccountID,
			UserID:      t.givenUserID,
			Constraints: t.given,
		}
		err = given.ValidateConstraints(tx)

		c.Assert(tx.Commit(), gc.IsNil)

		if t.expectError != "" {
			c.Check(err, gc.ErrorMatches, t.expectError)
			continue
		}

		c.Assert(err, gc.IsNil)

		c.Check(given.Constraints, gc.DeepEquals, t.expectConstraints)
	}
}

func (m *ModelSuite) TestSampleValidate(c *gc.C) {
	for i, t := range []struct {
		should   string
		given    model.Sample
		isInsert bool
		expect   errors.Errors
	}{{
		should: "give error for invalid Operator",
		given: model.Sample{
			Constraints: []stats.Constraint{
				{Key: "api.id", Operator: "foo"},
			},
		},
		isInsert: true,
		expect:   errors.Errors{"operator": {`invalid operator "foo"`}},
	}} {
		c.Logf("test %d: should %s", i, t.should)

		c.Check(t.given.Validate(t.isInsert), gc.DeepEquals, t.expect)
	}
}
