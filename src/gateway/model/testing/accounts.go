package testing

import (
	aperrors "gateway/errors"
	"gateway/model"
	apsql "gateway/sql"

	gc "gopkg.in/check.v1"
)

// Account fixtures.
const (
	JeffAccount  = "jeff"
	OtherAccount = "other"
)

// PrepareAccount adds the given account testing fixture to the given database.
func PrepareAccount(c *gc.C, db *apsql.DB, which string) *model.Account {
	tx, err := db.Begin()
	c.Assert(err, gc.IsNil)

	a, ok := accounts[which]
	c.Assert(ok, gc.Equals, true)
	acc := &a

	c.Assert(acc.Validate(true), gc.DeepEquals, make(aperrors.Errors))
	c.Assert(acc.Insert(tx), gc.IsNil)
	c.Assert(tx.Commit(), gc.IsNil)
	return acc
}

var accounts = map[string]model.Account{
	JeffAccount: {
		Name: `Jeff inc`,
	},
	OtherAccount: {
		Name: `Brian inc`,
	},
}
