package model

// User represents a user!
type User struct {
	Account  Account
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
