package testing

import (
	aperrors "gateway/errors"
	"gateway/model"
	apsql "gateway/sql"

	gc "gopkg.in/check.v1"
)

// API fixtures.
const (
	API1 = iota
	API2
	API3
)

// PrepareAPI adds the given account testing fixture to the given database.
func PrepareAPI(
	c *gc.C,
	db *apsql.DB,
	accID, userID int64,
	which int,
) *model.API {
	tx, err := db.Begin()
	c.Assert(err, gc.IsNil)
	a, ok := apis[which]
	c.Assert(ok, gc.Equals, true)
	a.AccountID = accID
	a.UserID = userID
	api := &a

	c.Assert(api.Validate(true), gc.DeepEquals, make(aperrors.Errors))
	c.Assert(api.Insert(tx), gc.IsNil)
	c.Assert(tx.Commit(), gc.IsNil)
	return api
}

var apis = map[int]model.API{
	API1: {
		Name:                 `Jeff API 1`,
		CORSAllowOrigin:      "*",
		CORSAllowHeaders:     "content-type, accept",
		CORSAllowCredentials: true,
		CORSRequestHeaders:   "*",
		CORSMaxAge:           600,
	},
	API2: {
		Name:                 `Jeff API 2`,
		CORSAllowOrigin:      "*",
		CORSAllowHeaders:     "content-type, accept",
		CORSAllowCredentials: true,
		CORSRequestHeaders:   "*",
		CORSMaxAge:           600,
	},
	API3: {
		Name:                 `Brian API 1`,
		CORSAllowOrigin:      "*",
		CORSAllowHeaders:     "content-type, accept",
		CORSAllowCredentials: true,
		CORSRequestHeaders:   "*",
		CORSMaxAge:           600,
	},
}
