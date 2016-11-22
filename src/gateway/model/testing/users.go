package testing

import (
	"fmt"
	aperrors "gateway/errors"
	"gateway/model"
	apsql "gateway/sql"
	"math/rand"
	"time"

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

	u, ok := users[which]
	c.Assert(ok, gc.Equals, true)
	u.AccountID = accID
	user := &u

	c.Assert(user.Validate(true), gc.DeepEquals, make(aperrors.Errors))
	c.Assert(user.Insert(tx), gc.IsNil)
	c.Assert(tx.Commit(), gc.IsNil)
	return user
}

var users = map[string]model.User{
	JeffUser: {
		Name:                    `Jeff`,
		Email:                   generateUniqueEmail("g%v@ffery.com"),
		NewPassword:             `password`,
		NewPasswordConfirmation: `password`,
		Admin:     true,
		Confirmed: true,
	},
	OtherUser: {
		Name:                    `Brian`,
		Email:                   generateUniqueEmail("br%v@in.com"),
		NewPassword:             `password`,
		NewPasswordConfirmation: `password`,
		Admin:     true,
		Confirmed: true,
	},
}

func generateUniqueEmail(formatter string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf(formatter, r.Int63())
}
