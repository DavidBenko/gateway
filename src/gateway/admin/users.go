package admin

import (
	"gateway/model"
	"net/http"
)

//go:generate ./controller.rb --model User --account --custom-struct --transform-method c.sanitize --transform-type sanitizedUser

// UsersController manages users.
type UsersController struct {
	BaseController
	accountID func(r *http.Request) int64
}

type sanitizedUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *UsersController) sanitize(user *model.User) *sanitizedUser {
	return &sanitizedUser{user.ID, user.Name, user.Email}
}
