package mail

import (
	"gateway/config"
	"gateway/model"
)

func SendWelcomeEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, resend, async bool) error {
	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "Welcome to JustAPIs!"
	context.Resend = resend
	err := Send("welcome.html", context, _smtp, user, async)
	if err != nil {
		return err
	}

	return nil
}
