package mail

import (
	"gateway/config"
	"gateway/model"
)

func SendInvoicePaymentFailedEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, async bool) error {
	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "Nanoscale.io Payment Failed"
	err := Send("payment_failure.html", context, _smtp, user, async)
	if err != nil {
		return err
	}

	return nil
}
