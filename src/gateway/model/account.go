package model

import (
	"errors"
	"fmt"

	aperrors "gateway/errors"
	"gateway/license"
	"gateway/sql"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
)

// Account represents a single tenant in multi-tenant deployment.
type Account struct {
	ID                        int64          `json:"id"`
	Name                      string         `json:"name"`
	PlanID                    sql.NullInt64  `json:"plan_id,omitempty" db:"plan_id"`
	StripeToken               string         `json:"stripe_token,omitempty" db:"-"`
	StripeCustomerID          sql.NullString `json:"-" db:"stripe_customer_id"`
	StripeSubscriptionID      sql.NullString `json:"-" db:"stripe_subscription_id"`
	StripePaymentRetryAttempt int64          `json:"-" db:"stripe_payment_retry_attempt"`
}

// Validate validates the model.
func (a *Account) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if a.Name == "" {
		errors.Add("name", "must not be blank")
	}
	// Check for Stripe integration expectations.
	if stripe.Key != "" {
		if a.PlanID.Int64 < 1 {
			errors.Add("plan", "must not be blank")
		}
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (a *Account) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if sql.IsUniqueConstraint(err, "accounts", "stripe_customer_id") {
		errors.Add("stripe_customer_id", "is already taken")
	}
	if sql.IsUniqueConstraint(err, "accounts", "stripe_subscription_id") {
		errors.Add("stripe_subscription_id", "is already taken")
	}
	return errors
}

// AllAccounts returns all accounts in default order.
func AllAccounts(db *sql.DB) ([]*Account, error) {
	accounts := []*Account{}
	err := db.Select(&accounts, db.SQL("accounts/all"))
	return accounts, err
}

// AllAccountsWithoutStripeCustomer returns all accounts without corresponding Stripe customer IDs in default order.
func AllAccountsWithoutStripeCustomer(db *sql.DB) ([]*Account, error) {
	accounts := []*Account{}
	err := db.Select(&accounts, db.SQL("accounts/all_without_stripe_customer"))
	return accounts, err
}

// MigrateAccountsToStripe creates corresponding customers in Stripe for every account in the database without a stripe_customer_id.
func MigrateAccountsToStripe(db *sql.DB, planName string) error {
	accounts, err := AllAccountsWithoutStripeCustomer(db)
	if err != nil {
		return err
	}
	defaultPlan, err := FindPlanByName(db, planName)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		account.PlanID.Int64 = defaultPlan.ID
		account.PlanID.Valid = true
		err = db.DoInTransaction(func(tx *sql.Tx) error {
			return account.Update(tx)
		})
		if err != nil {
			return err
		}
	}
	return err
}

// FirstAccount returns the first account found.
func FirstAccount(db *sql.DB) (*Account, error) {
	account := Account{}
	err := db.Get(&account, db.SQL("accounts/first"))
	return &account, err
}

// FindAccount returns the account with the id specified.
func FindAccount(db *sql.DB, id int64) (*Account, error) {
	account := Account{}
	err := db.Get(&account, db.SQL("accounts/find"), id)
	return &account, err
}

// FindAccountByStripeCustomer returns the account with the matching StripeCustomerID.
func FindAccountByStripeCustomer(db *sql.DB, customerID string) (*Account, error) {
	account := Account{}
	err := db.Get(&account, db.SQL("accounts/find_by_stripe_customer_id"), customerID)
	return &account, err
}

// DeleteAccount deletes the account with the id specified.
func DeleteAccount(tx *sql.Tx, id int64) error {
	currentAccount, accErr := FindAccount(tx.DB, id)
	err := tx.DeleteOne(tx.SQL("accounts/delete"), id)
	if err != nil {
		return err
	}
	if accErr == nil && stripe.Key != "" && currentAccount.StripeCustomerID.String != "" {
		_, err = customer.Del(currentAccount.StripeCustomerID.String)
		if err != nil {
			return err
		}
	}
	return tx.Notify("accounts", id, 0, 0, 0, id, sql.Delete)
}

// Insert inserts the account into the database as a new row.
func (a *Account) Insert(tx *sql.Tx) (err error) {
	if license.DeveloperVersion {
		var count int
		tx.Get(&count, tx.SQL("accounts/count"))
		if count >= license.DeveloperVersionAccounts {
			return errors.New(fmt.Sprintf("Developer version allows %v account(s).", license.DeveloperVersionAccounts))
		}
	}
	if stripe.Key != "" {
		plan, err := FindPlan(tx.DB, a.PlanID.Int64)
		if err != nil {
			return err
		}
		if plan.Price > 0 && a.StripeToken == "" {
			return errors.New("stripe_token must not be blank")
		}
		// Pass Stripe single-use token, plan, and customer details to Stripe to create the subscription.
		plan, err = FindPlan(tx.DB, a.PlanID.Int64)
		if err != nil {
			return err
		}
		customerParams := &stripe.CustomerParams{}
		if a.StripeToken != "" {
			customerParams.SetSource(a.StripeToken)
		}
		customerParams.Plan = plan.StripeName
		c, err := customer.New(customerParams)
		if err != nil {
			return err
		}
		a.ID, err = tx.InsertOne(tx.SQL("accounts/insert"), a.Name, a.PlanID, c.ID, c.Subs.Values[0].ID)
	} else {
		a.ID, err = tx.InsertOne(tx.SQL("accounts/insert"), a.Name, nil, nil, nil)
	}
	if err != nil {
		return err
	}
	return tx.Notify("accounts", a.ID, 0, 0, 0, a.ID, sql.Insert)
}

// Update updates the account in the database.
func (a *Account) Update(tx *sql.Tx) error {
	// Handle Stripe related plan or billing changes.
	if stripe.Key != "" {
		currentAccount, err := FindAccount(tx.DB, a.ID)
		if err != nil {
			return err
		}
		plan, err := FindPlan(tx.DB, a.PlanID.Int64)
		if err != nil {
			return err
		}
		if a.StripeCustomerID.String == "" {
			customerParams := &stripe.CustomerParams{}
			if a.StripeToken != "" {
				customerParams.SetSource(a.StripeToken)
			}
			customerParams.Plan = plan.StripeName
			c, err := customer.New(customerParams)
			if err != nil {
				return err
			}
			a.StripeCustomerID.String = c.ID
			a.StripeCustomerID.Valid = true
			a.StripeSubscriptionID.String = c.Subs.Values[0].ID
			a.StripeSubscriptionID.Valid = true
			a.PlanID.Int64 = plan.ID
			currentAccount.StripeCustomerID.String = c.ID
			currentAccount.StripeCustomerID.Valid = true
			currentAccount.StripeSubscriptionID.String = c.Subs.Values[0].ID
			currentAccount.StripeSubscriptionID.Valid = true
			currentAccount.PlanID.Int64 = plan.ID
			err = tx.UpdateOne(tx.SQL("accounts/update_stripe_customer_details"), c.ID, c.Subs.Values[0].ID, a.PlanID.Int64, a.ID)
			if err != nil {
				return err
			}
		}
		if a.PlanID.Int64 != currentAccount.PlanID.Int64 && plan.Price > 0 {
			c, err := customer.Get(currentAccount.StripeCustomerID.String, nil)
			if err != nil {
				return err
			}
			if c.DefaultSource == nil && a.StripeToken == "" {
				return errors.New("stripe_token must not be blank")
			}
		}
		if a.StripeToken != "" {
			customerParams := &stripe.CustomerParams{}
			customerParams.SetSource(a.StripeToken)
			_, err = customer.Update(currentAccount.StripeCustomerID.String, customerParams)
			if err != nil {
				return err
			}
		}
		if a.PlanID.Int64 != currentAccount.PlanID.Int64 {
			_, err = sub.Update(currentAccount.StripeSubscriptionID.String,
				&stripe.SubParams{
					Plan: plan.StripeName,
				},
			)
			if err != nil {
				return err
			}
		}
		err = tx.UpdateOne(tx.SQL("accounts/update"), a.Name, a.PlanID, a.ID)
	}
	err := tx.UpdateOne(tx.SQL("accounts/update"), a.Name, nil, a.ID)
	if err != nil {
		return err
	}
	return tx.Notify("accounts", a.ID, 0, 0, 0, a.ID, sql.Update)
}

// SetStripePaymentRetryAttempt updates the account in the database with a new StripePaymentRetryAttempt value.
func (a *Account) SetStripePaymentRetryAttempt(tx *sql.Tx, retry int64) error {
	a.StripePaymentRetryAttempt = retry
	return tx.UpdateOne(tx.SQL("accounts/update_stripe_payment_retry_attempt"), retry, a.ID)
}
