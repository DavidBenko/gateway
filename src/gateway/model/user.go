package model

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	aperrors "gateway/errors"
	aphttp "gateway/http"
	"gateway/license"
	apsql "gateway/sql"

	"github.com/stripe/stripe-go"
	"golang.org/x/crypto/bcrypt"
)

const bcryptPasswordCost = 10

var symbols = []byte("0123456789ABCDEF")
var isymbols = map[byte]int64{}

type TokenType int

const (
	TokenTypeReset TokenType = iota
	TokenTypeConfirm
)

func init() {
	for i, j := range symbols {
		isymbols[j] = int64(i)
	}
}

// User represents a user!
type User struct {
	AccountID               int64  `json:"-" db:"account_id"`
	UserID                  int64  `json:"-"`
	ID                      int64  `json:"id"`
	Name                    string `json:"name"`
	Email                   string `json:"email"`
	Admin                   bool   `json:"admin"`
	Token                   string `json:"token"`
	Confirmed               bool   `json:"confirmed"`
	NewPassword             string `json:"password"`
	NewPasswordConfirmation string `json:"password_confirmation"`
	HashedPassword          string `json:"-" db:"hashed_password"`
}

// Validate validates the model.
func (u *User) Validate(isInsert bool) aperrors.Errors {
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

func (u *User) HasConfirmToken() bool {
	return strings.HasPrefix(u.Token, "confirm-")
}

// AllUsersForAccountID returns all users on the Account in default order.
func AllUsersForAccountID(db *apsql.DB, accountID int64) ([]*User, error) {
	users := []*User{}
	err := db.Select(&users,
		`SELECT id, name, email, admin, confirmed FROM users
		 WHERE account_id = ? ORDER BY name ASC;`,
		accountID)
	return users, err
}

// FindUserForAccountID returns the user with the id and account_id specified.
func FindUserForAccountID(db *apsql.DB, id, accountID int64) (*User, error) {
	user := User{}
	err := db.Get(&user,
		`SELECT id, name, email, admin, confirmed FROM users
		 WHERE id = ? AND account_id = ?;`,
		id, accountID)
	return &user, err
}

// FindFirstUserForAccountID returns the first user with the account_id specified.
func FindFirstUserForAccountID(db *apsql.DB, accountID int64) (*User, error) {
	user := User{}
	err := db.Get(&user,
		`SELECT id, name, email, admin, confirmed FROM users
		 WHERE account_id = ? ORDER BY id LIMIT 1;`,
		accountID)
	return &user, err
}

// FindAdminUserForAccountID returns the first admin user with the account_id specified.
func FindAdminUserForAccountID(db *apsql.DB, accountID int64) (*User, error) {
	user := User{}
	err := db.Get(&user,
		`SELECT id, name, email, admin, confirmed FROM users
		 WHERE account_id = ? AND admin = 1 ORDER BY id LIMIT 1;`,
		accountID)
	return &user, err
}

func CanDeleteUser(tx *apsql.Tx, id, accountID int64, auth aphttp.AuthType) error {
	if auth == aphttp.AuthTypeSite {
		return nil
	}

	user := User{}
	err := tx.Get(&user, tx.SQL("users/find"), id, accountID)
	if err != nil {
		return apsql.ErrZeroRowsAffected
	}

	var count int
	err = tx.Get(&count, tx.SQL("users/count_admin"), accountID, true)
	if err != nil {
		return nil
	}

	if count == 1 {
		if user.Admin {
			return errors.New("There must be at least one admin user")
		}
	}

	return nil
}

// DeleteUserForAccountID deletes the user with the id and account_id specified.
func DeleteUserForAccountID(tx *apsql.Tx, id, accountID, userID int64) error {
	err := tx.DeleteOne(
		`DELETE FROM users
		 WHERE id = ? AND account_id = ?;`,
		id, accountID)
	if err != nil {
		return err
	}

	return tx.Notify("users", accountID, userID, 0, 0, id, apsql.Delete)
}

// FindUserByEmail returns the user with the email specified.
func FindUserByEmail(db *apsql.DB, email string) (*User, error) {
	user := User{}
	err := db.Get(&user,
		`SELECT id, account_id, name, email, admin, token, confirmed, hashed_password
		 FROM users WHERE email = ?;`,
		strings.ToLower(email))
	return &user, err
}

func AddUserToken(tx *apsql.Tx, email string, tokenType TokenType) (string, error) {
	buffer := make([]byte, 32)
	timestamp := time.Now().Unix()
	for i := range buffer[:16] {
		buffer[i] = symbols[timestamp&0xF]
		timestamp >>= 4
	}
	for i := range buffer[16:] {
		buffer[i+16] = symbols[rand.Intn(len(symbols))]
	}

	token := string(buffer)
	switch tokenType {
	case TokenTypeReset:
		token = "reset-" + token
	case TokenTypeConfirm:
		token = "confirm-" + token
	default:
		return "", errors.New("invalid token type")
	}
	err := tx.UpdateOne(tx.SQL("users/add_token"), token, email)
	if err != nil {
		return "", err
	}

	return token, nil
}

func ValidateUserToken(tx *apsql.Tx, token string, delete bool) (*User, error) {
	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		return nil, errors.New("token must be of the form 'type-123ABC...'")
	}

	var tokenType TokenType
	switch parts[0] {
	case "reset":
		tokenType = TokenTypeReset
	case "confirm":
		tokenType = TokenTypeConfirm
	default:
		return nil, errors.New("invalid token type")
	}

	if len(parts[1]) != 32 {
		return nil, errors.New("token code must be 32 bytes long")
	}

	user := User{}
	err := tx.Get(&user, tx.SQL("users/find_token"), token)
	if err != nil {
		return nil, err
	}
	if tokenType == TokenTypeReset {
		var unix int64
		for i, b := range []byte(parts[1])[:16] {
			unix |= isymbols[b] << (4 * uint(i))
		}
		if time.Unix(unix, 0).Add(24 * time.Hour).Before(time.Now()) {
			return nil, errors.New("token timestamp is stale")
		}
	}
	if delete {
		err = tx.UpdateOne(tx.SQL("users/add_token"), "", user.Email)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}

// FindUserByID returns the user with the id specified.
func FindUserByID(db *apsql.DB, id int64) (*User, error) {
	user := User{}
	err := db.Get(&user, db.SQL("users/find_id"), id)
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

	if stripe.Key != "" {
		account, err := FindAccount(tx.DB, u.AccountID)
		if err != nil {
			return err
		}
		plan, err := FindPlan(tx.DB, account.PlanID.Int64)
		if err != nil {
			return err
		}
		if int64(count) >= plan.MaxUsers {
			return errors.New(fmt.Sprintf("Plan only allows %v user(s).", plan.MaxUsers))
		}
	}

	if count == 0 {
		u.Admin = true
	}

	if err = u.hashPassword(); err != nil {
		return err
	}

	u.ID, err = tx.InsertOne(
		`INSERT INTO users (account_id, name, email, admin, token, confirmed, hashed_password)
		 VALUES (?, ?, ?, ?, '', ?, ?)`,
		u.AccountID, u.Name, strings.ToLower(u.Email), u.Admin, u.Confirmed, u.HashedPassword)
	if err != nil {
		return err
	}

	return tx.Notify("users", u.AccountID, u.UserID, 0, 0, u.ID, apsql.Insert)
}

// Update updates the user in the database.
func (u *User) Update(tx *apsql.Tx) error {
	var err error
	var count int
	err = tx.Get(&count, tx.SQL("users/count_admin"), u.AccountID, true)
	if err != nil {
		return apsql.ErrZeroRowsAffected
	}

	user := User{}
	err = tx.Get(&user, tx.SQL("users/find_id"), u.ID)
	if err != nil {
		return apsql.ErrZeroRowsAffected
	}
	if user.Confirmed && !u.Confirmed {
		return errors.New("Can't unconfirm user")
	}
	if count == 1 && user.Admin && !u.Admin {
		return errors.New("There must be at least one admin user")
	}

	if u.NewPassword != "" {
		err = u.hashPassword()
		if err != nil {
			return err
		}
		err = tx.UpdateOne(
			`UPDATE users
			 SET name = ?, email = ?, admin = ?, confirmed = ?, hashed_password = ?
			 WHERE id = ? AND account_id = ?;`,
			u.Name, strings.ToLower(u.Email), u.Admin, u.Confirmed, u.HashedPassword, u.ID, u.AccountID)
		if err != nil {
			return err
		}

		return tx.Notify("users", u.AccountID, u.UserID, 0, 0, u.ID, apsql.Update)
	}

	err = tx.UpdateOne(
		`UPDATE users
			 SET name = ?, email = ?, admin = ?, confirmed = ?
			 WHERE id = ? AND account_id = ?;`,
		u.Name, strings.ToLower(u.Email), u.Admin, u.Confirmed, u.ID, u.AccountID)
	if err != nil {
		return err
	}

	return tx.Notify("users", u.AccountID, u.UserID, 0, 0, u.ID, apsql.Update)
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
