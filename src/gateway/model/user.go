package model

import (
	"errors"
	"fmt"
	"strings"

	aperrors "gateway/errors"
	"gateway/license"
	apsql "gateway/sql"

	"golang.org/x/crypto/bcrypt"
)

const bcryptPasswordCost = 10

// User represents a user!
type User struct {
	AccountID               int64  `json:"-" db:"account_id"`
	UserID                  int64  `json:"-"`
	ID                      int64  `json:"id"`
	Name                    string `json:"name"`
	Email                   string `json:"email"`
	Admin                   bool   `json:"admin"`
	NewPassword             string `json:"password"`
	NewPasswordConfirmation string `json:"password_confirmation"`
	HashedPassword          string `json:"-" db:"hashed_password"`
}

// Validate validates the model.
func (u *User) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)
	if u.Name == "" {
		errors.Add("name", "must not be blank")
	}
	if u.Email == "" {
		errors.Add("email", "must not be blank")
	}
	if u.ID == 0 && u.NewPassword == "" {
		errors.Add("password", "must not be blank")
	}
	if u.NewPassword != "" && (u.NewPassword != u.NewPasswordConfirmation) {
		errors.Add("password_confirmation", "must match password")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (u *User) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if err.Error() == "UNIQUE constraint failed: users.email" ||
		err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
		errors.Add("email", "is already taken")
	}
	return errors
}

// AllUsersForAccountID returns all users on the Account in default order.
func AllUsersForAccountID(db *apsql.DB, accountID int64) ([]*User, error) {
	users := []*User{}
	err := db.Select(&users,
		`SELECT id, name, email, admin FROM users
		 WHERE account_id = ? ORDER BY name ASC;`,
		accountID)
	return users, err
}

// FindUserForAccountID returns the user with the id and account_id specified.
func FindUserForAccountID(db *apsql.DB, id, accountID int64) (*User, error) {
	user := User{}
	err := db.Get(&user,
		`SELECT id, name, email, admin FROM users
		 WHERE id = ? AND account_id = ?;`,
		id, accountID)
	return &user, err
}

// FindFirstUserForAccountID returns the first user with the account_id specified.
func FindFirstUserForAccountID(db *apsql.DB, accountID int64) (*User, error) {
	user := User{}
	err := db.Get(&user,
		`SELECT id, name, email, admin FROM users
		 WHERE account_id = ? ORDER BY id LIMIT 1;`,
		accountID)
	return &user, err
}

// DeleteUserForAccountID deletes the user with the id and account_id specified.
func DeleteUserForAccountID(tx *apsql.Tx, id, accountID, userID int64) error {
	return tx.DeleteOne(
		`DELETE FROM users
		 WHERE id = ? AND account_id = ?;`,
		id, accountID)
}

// FindUserByEmail returns the user with the email specified.
func FindUserByEmail(db *apsql.DB, email string) (*User, error) {
	user := User{}
	err := db.Get(&user,
		`SELECT id, account_id, admin, hashed_password
		 FROM users WHERE email = ?;`,
		strings.ToLower(email))
	return &user, err
}

// FindUserByID returns the user with the id specified.
func FindUserByID(db *apsql.DB, id int64) (*User, error) {
	user := User{}
	err := db.Get(&user,
		`SELECT id, account_id, name, email, admin
		 FROM users WHERE id = ?;`,
		id)
	return &user, err
}

// Insert inserts the user into the database as a new row.
func (u *User) Insert(tx *apsql.Tx) (err error) {
	var count int
	tx.Get(&count, tx.SQL("users/count"), u.AccountID)

	if license.DeveloperVersion {
		if count >= license.DeveloperVersionUsers {
			return errors.New(fmt.Sprintf("Developer version allows %v user(s).", license.DeveloperVersionUsers))
		}
	}

	if count == 0 {
		u.Admin = true
	}

	if err = u.hashPassword(); err != nil {
		return err
	}

	u.ID, err = tx.InsertOne(
		`INSERT INTO users (account_id, name, email, admin, hashed_password)
		 VALUES (?, ?, ?, ?, ?)`,
		u.AccountID, u.Name, strings.ToLower(u.Email), u.Admin, u.HashedPassword)
	return err
}

// Update updates the user in the database.
func (u *User) Update(tx *apsql.Tx) error {
	var count int
	tx.Get(&count, tx.SQL("users/count"), u.AccountID)

	if count == 1 {
		u.Admin = true
	}

	var err error
	if u.NewPassword != "" {
		err = u.hashPassword()
		if err != nil {
			return err
		}
		return tx.UpdateOne(
			`UPDATE users
			 SET name = ?, email = ?, admin = ?, hashed_password = ?
			 WHERE id = ? AND account_id = ?;`,
			u.Name, strings.ToLower(u.Email), u.Admin, u.HashedPassword, u.ID, u.AccountID)
	}

	return tx.UpdateOne(
		`UPDATE users
			 SET name = ?, email = ?, admin = ?
			 WHERE id = ? AND account_id = ?;`,
		u.Name, strings.ToLower(u.Email), u.Admin, u.ID, u.AccountID)
}

func (u *User) hashPassword() error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.NewPassword), bcryptPasswordCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashed)
	return nil
}

// ValidPassword returns whether or not the password matches what's on file.
func (u *User) ValidPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}
