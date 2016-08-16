package mail

import (
	"gateway/config"
	"gateway/model"
)

func SendInvoicePaymentFailedAndPlanDowngradedEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, async bool) error {
	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "Payment Failed and Plan Downgraded"
	err := Send("payment_failure_downgrade.html", context, _smtp, user, async)
	if err != nil {
		return err
	}

	return nil
}
