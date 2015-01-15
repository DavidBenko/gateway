package model

import (
	"database/sql"
	"fmt"
	"gateway/config"
	apsql "gateway/sql"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

const bcryptPasswordCost = 10

// User represents a user!
type User struct {
	AccountID               int64  `json:"-" db:"account_id"`
	ID                      int64  `json:"id"`
	Name                    string `json:"name"`
	Email                   string `json:"email"`
	NewPassword             string `json:"password"`
	NewPasswordConfirmation string `json:"password_confirmation"`
	HashedPassword          string `json:"-" db:"hashed_password"`
}

// Validate validates the model.
func (u *User) Validate() Errors {
	errors := make(Errors)
	if u.Name == "" {
		errors.add("name", "must not be blank")
	}
	if u.Email == "" {
		errors.add("email", "must not be blank")
	}
	if u.ID == 0 && u.NewPassword == "" {
		errors.add("password", "must not be blank")
	}
	if u.NewPassword != "" && (u.NewPassword != u.NewPasswordConfirmation) {
		errors.add("password_confirmation", "must match password")
	}
	return errors
}

// AllUsersForAccountID returns all users on the Account in default order.
func AllUsersForAccountID(db *apsql.DB, accountID int64) ([]*User, error) {
	users := []*User{}
	err := db.Select(&users,
		"SELECT `id`, `name`, `email` FROM `users` WHERE account_id = ? ORDER BY `name` ASC;",
		accountID)
	return users, err
}

// FindUserForAccountID returns the user with the id and account_id specified.
func FindUserForAccountID(db *apsql.DB, id, accountID int64) (*User, error) {
	user := User{}
	err := db.Get(&user, "SELECT `id`, `name`, `email` FROM `users` WHERE `id` = ? AND account_id = ?;",
		id, accountID)
	return &user, err
}

// DeleteUserForAccountID deletes the user with the id and account_id specified.
func DeleteUserForAccountID(tx *sqlx.Tx, id, accountID int64) error {
	result, err := tx.Exec("DELETE FROM `users` WHERE `id` = ? AND account_id = ?;",
		id, accountID)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}

	return nil
}

// FindUserByEmail returns the user with the email specified.
func FindUserByEmail(db *apsql.DB, email string) (*User, error) {
	user := User{}
	err := db.Get(&user,
		"SELECT `id`, `account_id`, `hashed_password` FROM `users` WHERE `email` = ?;",
		strings.ToLower(email))
	return &user, err
}

// Insert inserts the user into the database as a new row.
func (u *User) Insert(tx *sqlx.Tx) error {
	err := u.hashPassword()
	if err != nil {
		return err
	}

	result, err := tx.Exec("INSERT INTO `users` (`account_id`, `name`, `email`, `hashed_password`) VALUES (?, ?, ?, ?);",
		u.AccountID, u.Name, strings.ToLower(u.Email), u.HashedPassword)
	if err != nil {
		return err
	}
	u.ID, err = result.LastInsertId()
	if err != nil {
		log.Printf("%s Error getting last insert ID for user: %v",
			config.System, err)
		return err
	}
	return nil
}

// Update updates the user in the database.
func (u *User) Update(tx *sqlx.Tx) error {
	var result sql.Result
	var err error
	if u.NewPassword != "" {
		err = u.hashPassword()
		if err != nil {
			return err
		}
		result, err = tx.Exec("UPDATE `users` SET `name` = ?, `email` = ?, `hashed_password` = ? WHERE `id` = ? AND `account_id` = ?;",
			u.Name, strings.ToLower(u.Email), u.HashedPassword, u.ID, u.AccountID)
	} else {
		result, err = tx.Exec("UPDATE `users` SET `name` = ?, `email` = ? WHERE `id` = ? AND `account_id` = ?;",
			u.Name, strings.ToLower(u.Email), u.ID, u.AccountID)
	}
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}
	return nil
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
