package testing

import (
	aperrors "gateway/errors"
	"gateway/model"
	apsql "gateway/sql"

	gc "gopkg.in/check.v1"
)

// User fixtures.
const (
	JeffUser  = "jeff"
	OtherUser = "other"
)

// PrepareUser adds the given account testing fixture to the given database.
func PrepareUser(
	c *gc.C,
	db *apsql.DB,
	accID int64,
	which string,
) *model.User {
	tx, err := db.Begin()
	c.Assert(err, gc.IsNil)
	defer func() { c.Assert(tx.Commit(), gc.IsNil) }()

	u, ok := users[which]
	c.Assert(ok, gc.Equals, true)
	u.AccountID = accID
	user := &u

	c.Assert(user.Validate(true), gc.DeepEquals, make(aperrors.Errors))
	c.Assert(user.Insert(tx), gc.IsNil)

	return user
}

var users = map[string]model.User{
	JeffUser: {
		Name:                    `Jeff`,
		Email:                   `g@ffery.com`,
		NewPassword:             `password`,
		NewPasswordConfirmation: `password`,
		Admin:     true,
		Confirmed: true,
	},
	OtherUser: {
		Name:                    `Brian`,
		Email:                   `br@in.com`,
		NewPassword:             `password`,
		NewPasswordConfirmation: `password`,
		Admin:     true,
		Confirmed: true,
	},
}
