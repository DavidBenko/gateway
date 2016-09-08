package admin

import (
	"errors"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"
	"io"
	"net/http"
)

// AccountsController manages Accounts.
type AccountController struct {
	BaseController
}

// Show shows the Account.
func (c *AccountController) Show(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	user, err := model.FindUserByID(db, c.userID(r))
	if err != nil {
		return aphttp.NewServerError(err)
	}
	if !user.Admin {
		return c.forbidden()
	}
	account, err := model.FindAccount(db, c.accountID(r))

	if err != nil {
		return c.notFound()
	}

	return c.serializeInstance(account, w)
}

// Create creates the Account.
func (c *AccountController) Create(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {
	return c.forbidden()
}

// Update updates the Account.
func (c *AccountController) Update(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	user, err := model.FindUserByID(tx.DB, c.userID(r))
	if err != nil {
		return aphttp.NewServerError(err)
	}
	if !user.Admin {
		return c.forbidden()
	}

	account, httpErr := c.deserializeInstance(r.Body)
	if httpErr != nil {
		return httpErr
	}

	account.ID = c.accountID(r)

	var method func(*apsql.Tx) error
	desc := "updating"
	method = account.Update

	validationErrors := account.Validate(false)
	if !validationErrors.Empty() {
		return SerializableValidationErrors{validationErrors}
	}

	if err := method(tx); err != nil {
		if err == apsql.ErrZeroRowsAffected {
			return c.notFound()
		}
		validationErrors = account.ValidateFromDatabaseError(err)
		if !validationErrors.Empty() {
			return SerializableValidationErrors{validationErrors}
		} else {
			validationErrors = account.ValidateFromStripeError(err)
			if !validationErrors.Empty() {
				return SerializableValidationErrors{validationErrors}
			}
		}
		logreport.Printf("%s Error %s account: %v\n%v", config.System, desc, err, r)
		return aphttp.NewServerError(err)
	}

	return c.serializeInstance(account, w)
}

// Delete deletes the Acount.
func (c *AccountController) Delete(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {
	return c.forbidden()
}

func (c *AccountController) mapFields(r *http.Request, object *model.Account) {
	if c.accountID != nil {
		mapAccountID(c.accountID(r), object)
	}
	if c.userID != nil {
		mapUserID(c.userID(r), object)
	}
	mapFromPath(r, object)
}

func (c *AccountController) notFound() aphttp.Error {
	return aphttp.NewError(errors.New("No account matches"), 404)
}

func (c *AccountController) forbidden() aphttp.Error {
	return aphttp.NewError(errors.New("Forbidden"), 403)
}

func (c *AccountController) deserializeInstance(file io.Reader) (*model.Account,
	aphttp.Error) {

	var wrapped struct {
		Account *model.Account `json:"account"`
	}
	if err := deserialize(&wrapped, file); err != nil {
		return nil, err
	}
	if wrapped.Account == nil {
		return nil, aphttp.NewError(errors.New("Could not deserialize Account from JSON."),
			http.StatusBadRequest)
	}
	return wrapped.Account, nil
}

func (c *AccountController) serializeInstance(instance *model.Account,
	w http.ResponseWriter) aphttp.Error {

	wrapped := struct {
		Account *model.Account `json:"account"`
	}{instance}
	return serialize(wrapped, w)
}
