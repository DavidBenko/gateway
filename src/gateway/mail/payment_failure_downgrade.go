package mail

import (
	"gateway/config"
	"gateway/model"
)

func SendInvoicePaymentFailedAndPlanDowngradedEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, paymentDetails *PaymentDetails, async bool) error {
	context := NewEmailTemplateWithPaymentDetails(_smtp, proxyServer, admin, user, paymentDetails)
	context.Subject = "Nanoscale.io Billing Problem – Final Notice"
	err := Send("payment_failure_downgrade.html", context, _smtp, user, async)
	if err != nil {
		return err
	}

	return nil
}
