package model

import (
	"database/sql"
	"fmt"
	"gateway/config"
	apsql "gateway/sql"
	"log"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user!
type User struct {
	AccountID               int64  `json:"-"`
	ID                      int64  `json:"id"`
	Name                    string `json:"name"`
	Email                   string `json:"email"`
	NewPassword             string `json:"password"`
	NewPasswordConfirmation string `json:"password_confirmation"`
	HashedPassword          string `json:"-"`
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

// Insert inserts the account into the database as a new row.
func (u *User) Insert(tx *sqlx.Tx) error {
	err := u.hashPassword()
	if err != nil {
		return err
	}

	result, err := tx.Exec("INSERT INTO `users` (`account_id`, `name`, `email`, `hashed_password`) VALUES (?, ?, ?, ?);",
		u.AccountID, u.Name, u.Email, u.HashedPassword)
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

// Update updates the account in the database.
func (u *User) Update(tx *sqlx.Tx) error {
	var result sql.Result
	var err error
	if u.NewPassword != "" {
		err = u.hashPassword()
		if err != nil {
			return err
		}
		result, err = tx.Exec("UPDATE `users` SET `name` = ?, `email` = ?, `hashed_password` = ? WHERE `id` = ?;",
			u.Name, u.Email, u.HashedPassword, u.ID)
	} else {
		result, err = tx.Exec("UPDATE `users` SET `name` = ?, `email` = ? WHERE `id` = ?;",
			u.Name, u.Email, u.ID)
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
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.NewPassword), 10)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashed)
	return nil
}
