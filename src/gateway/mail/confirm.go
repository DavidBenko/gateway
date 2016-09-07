package mail

import (
	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

func SendConfirmEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, tx *apsql.Tx, async bool) error {
	token, resend := user.Token, true
	if token == "" {
		var err error
		token, err = model.AddUserToken(tx, user.Email, model.TokenTypeConfirm)
		if err != nil {
			return err
		}
		resend = false
	}

	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "Nanoscale.io Email Confirmation"
	context.Token = token
	context.Resend = resend
	err := Send("confirm.html", context, _smtp, user, async)
	if err != nil {
		return err
	}

	return nil
}
