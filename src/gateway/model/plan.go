package model

import (
	aperrors "gateway/errors"
	"gateway/sql"

	stripe "github.com/stripe/stripe-go"
)

// Plan represents a set of usage rules for an account.
type Plan struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	StripeName        string `json:"-" db:"stripe_name"`
	MaxUsers          int64  `json:"max_users" db:"max_users"`
	JavascriptTimeout int64  `json:"javascript_timeout" db:"javascript_timeout"`
	JobTimeout        int64  `json:"job_timeout" db:"job_timeout"`
	Price             int64  `json:"price"`
}

// Validate validates the model.
func (p *Plan) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if stripe.Key == "" {
		errors.Add("stripe_name", "stripe must be configured with API keys to use plans")
		return errors
	}
	if p.Name == "" {
		errors.Add("name", "must not be blank")
	}
	if p.StripeName == "" {
		errors.Add("stripe_name", "must not be blank")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (p *Plan) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if sql.IsUniqueConstraint(err, "plans", "name") {
		errors.Add("name", "is already taken")
	}
	if sql.IsUniqueConstraint(err, "plan", "stripe_name") {
		errors.Add("stripe_name", "is already taken")
	}
	return errors
}

// AllPlans returns all plans in default order.
func AllPlans(db *sql.DB) ([]*Plan, error) {
	plans := []*Plan{}
	err := db.Select(&plans, db.SQL("plans/all"))
	return plans, err
}

// FindPlan returns the plan with the id specified.
func FindPlan(db *sql.DB, id int64) (*Plan, error) {
	plan := Plan{}
	err := db.Get(&plan, db.SQL("plans/find"), id)
	return &plan, err
}

// FindPlanByName returns the plan with the name specified.
func FindPlanByName(db *sql.DB, name string) (*Plan, error) {
	plan := Plan{}
	err := db.Get(&plan, db.SQL("plans/find_by_name"), name)
	return &plan, err
}

// FindPlanByAccountID returns the plan that is associated with the specified AccountID.
func FindPlanByAccountID(db *sql.DB, accountID int64) (*Plan, error) {
	plan := Plan{}
	err := db.Get(&plan, db.SQL("plans/find_by_account_id"), accountID)
	return &plan, err
}

// Insert the plan into the database as a new row.
func (p *Plan) Insert(tx *sql.Tx) (err error) {
	p.ID, err = tx.InsertOne(tx.SQL("plans/insert"), p.Name, p.StripeName, p.MaxUsers, p.JavascriptTimeout, p.JobTimeout, p.Price)
	if err != nil {
		return err
	}
	return tx.Notify("plans", 0, 0, 0, 0, p.ID, sql.Insert)
}

// Update updates the plan in the database.
func (p *Plan) Update(tx *sql.Tx) error {
	err := tx.UpdateOne(tx.SQL("plans/update"), p.Name, p.StripeName, p.MaxUsers, p.JavascriptTimeout, p.JobTimeout, p.Price, p.ID)
	if err != nil {
		return err
	}
	return tx.Notify("plans", 0, 0, 0, 0, p.ID, sql.Update)
}

// DeletePlan deletes the plan with the id specified.
func DeletePlan(tx *sql.Tx, id int64) error {
	err := tx.DeleteOne(tx.SQL("plans/delete"), id)
	if err != nil {
		return err
	}
	return tx.Notify("plans", 0, 0, 0, 0, id, sql.Delete)
}
