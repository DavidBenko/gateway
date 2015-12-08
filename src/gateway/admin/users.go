package admin

import (
	"gateway/mail"
	"gateway/model"
	apsql "gateway/sql"
)

type sanitizedUser struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Admin     bool   `json:"admin"`
	Confirmed bool   `json:"confirmed"`
}

func (c *UsersController) sanitize(user *model.User) *sanitizedUser {
	return &sanitizedUser{user.ID, user.Name, user.Email, user.Admin, user.Confirmed}
}

func (c *UsersController) AfterInsert(user *model.User, tx *apsql.Tx) error {
	if user.Confirmed {
		return nil
	}

	return mail.SendConfirmEmail(c.SMTP, c.ProxyServer, c.conf, user, tx)
}
