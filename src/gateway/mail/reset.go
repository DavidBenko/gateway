package mail

import (
	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

func SendResetEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, tx *apsql.Tx, async bool) error {
	token, err := model.AddUserToken(tx, user.Email, model.TokenTypeReset)
	if err != nil {
		return err
	}

	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "JustAPIs Password Reset"
	context.Token = token
	err = Send("reset.html", context, _smtp, user, async)
	if err != nil {
		return err
	}

	return nil
}
