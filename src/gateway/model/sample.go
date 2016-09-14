package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	aperrors "gateway/errors"
	apsql "gateway/sql"
	"gateway/stats"
)

// Sample represents a query for a stats.Sampler.
type Sample struct {
	ID int64 `json:"id,omitempty"`

	Name string `json:"name"`

	// Variables are the values the sample will select.
	Variables []string `json:"variables,omitempty"`

	// Constraints are the constraints (WHERE foo OP bar) the sample will be
	// restricted on.
	Constraints []stats.Constraint `json:"constraints,omitempty"`

	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`
}

// We have to make sure:
//  - Queries restricting on api.id outside of own user are rejected.
//  - Query only returns data for the particular user.  (I.e. must
//    inject a new constraint on api.id's belonging to the given user.id.)

// ValidateConstraints validates the given constraints against the user's
// privileges.
func (s *Sample) ValidateConstraints(tx *apsql.Tx) error {
	var ownedAPIs []int64
	err := tx.Select(&ownedAPIs, `
  SELECT DISTINCT a.id
    FROM users u, apis a
   WHERE u.account_id = a.account_id
	   AND u.id = ?
  	 AND u.account_id = ?
ORDER BY a.id ASC`, s.UserID, s.AccountID)
	if err != nil {
		return err
	}

	owned := make(map[int64]bool)
	for _, id := range ownedAPIs {
		owned[id] = true
	}

	for _, c := range s.Constraints {
		if c.Key == `api.id` {
			// If the user wants to know something about api.id, the
			// user must be authorized.  First, make sure the value
			// and operator are correctly typed.
			switch tV := c.Value.(type) {
			case int64:
				if c.Operator != stats.EQ {
					return fmt.Errorf("invalid operator %q"+
						" for single api.id value,"+
						` use "EQ"`,
						c.Operator,
					)
				}
				if !owned[tV] {
					return fmt.Errorf(
						"api %d not owned by user %d",
						tV, s.UserID,
					)
				}
				return nil
			case []int64:
				if c.Operator != stats.IN {
					return fmt.Errorf("invalid operator %q"+
						" for multiple api.id values, "+
						`use "IN"`,
						c.Operator,
					)
				}
				var notOwned []string
				for _, id := range tV {
					if !owned[id] {
						notOwned = append(notOwned,
							strconv.Itoa(int(id)))
					}
				}
				var msg string
				switch len(notOwned) {
				case 0:
					return nil
				case 1:
					msg = fmt.Sprintf(
						"api %s not owned by user %d",
						notOwned[0],
						s.UserID,
					)
				default:
					msg = fmt.Sprintf(
						"apis %s not owned by user %d",
						strings.Join(notOwned, ", "),
						s.UserID,
					)
				}
				return errors.New(msg)
			default:
				return fmt.Errorf("invalid type %T for api.id"+
					" value, must be int64 or []int64", tV)
			}
		}
	}

	// If api.id wasn't specified in the query, it must be
	// constrained to the set of APIs owned by the user.
	s.Constraints = append(s.Constraints, stats.Constraint{
		Key:      "api.id",
		Operator: stats.IN,
		Value:    []int64(ownedAPIs),
	})

	return nil
}

// Validate validates the given Sample.
func (s *Sample) Validate(isInsert bool) aperrors.Errors {
	errs := make(aperrors.Errors)
	for _, c := range s.Constraints {
		errs.AddAll(c.Validate())
	}
	return errs
}

// ValidateFromDatabaseError validates.
func (s *Sample) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	// if apsql.IsUniqueConstraint(err, "proxy_endpoints", "api_id", "name") {
	// 	errors.Add("name", "is already taken")
	// }
	// if apsql.IsNotNullConstraint(err, "proxy_endpoints", "environment_id") {
	// 	errors.Add("environment_id", "must be a valid environment in this API")
	// }
	// if apsql.IsNotNullConstraint(err, "proxy_endpoint_calls", "remote_endpoint_id") {
	// 	errors.Add("components", "all calls must reference a valid remote endpoint in this API")
	// }
	// if apsql.IsUniqueConstraint(err, "proxy_endpoint_tests", "endpoint_id", "name") {
	// 	errors.Add("tests", "name is already taken")
	// }
	return errors
}

// Insert inserts a single Sample using the given Tx.
func (s *Sample) Insert(tx *apsql.Tx) error {
	return errors.New("implement me")
}

// Update updates a single Sample using the given Tx.
func (s *Sample) Update(tx *apsql.Tx) error {
	return errors.New("implement me")
}

// DeleteSampleForAccountID deletes the given sample.
func DeleteSampleForAccountID(
	tx *apsql.Tx,
	id, accountID, userID int64,
) error {
	return errors.New("implement me")
}

// AllSamplesForAccountID returns all samples for the given IDs.
func AllSamplesForAccountID(
	db *apsql.DB,
	accountID int64,
) ([]*Sample, error) {
	return nil, errors.New("implement me")
}

// FindSampleForAccountID finds the given sample.
func FindSampleForAccountID(
	db *apsql.DB,
	id, accountID int64,
) (*Sample, error) {
	return nil, errors.New("implement me")
}
