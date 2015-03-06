package admin

import "gateway/model"

type sanitizedUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *UsersController) sanitize(user *model.User) *sanitizedUser {
	return &sanitizedUser{user.ID, user.Name, user.Email}
}
