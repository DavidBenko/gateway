package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
	"time"

	"gateway/config"
	"gateway/logreport"
	"gateway/model"
)

type EmailTemplate struct {
	From           string
	To             string
	Subject        string
	Scheme         string
	Host           string
	Port           int64
	Prefix         string
	Token          string
	Resend         bool
	PaymentDetails *PaymentDetails
}

type PaymentDetails struct {
	InvoiceID         string
	PaymentAmount     uint64
	PaymentDate       int64
	Plan              string
	PlanAmount        uint64
	CardDisplay       string
	FailureReason     string
	NextChargeAttempt int64
}

func (p *PaymentDetails) PaymentDateString() string {
	t := time.Unix(p.PaymentDate, 0)
	return fmt.Sprintf("%s %d, %d", t.Month(), t.Day(), t.Year())
}

func (p *PaymentDetails) NextChargeAttemptString() string {
	t := time.Unix(p.NextChargeAttempt, 0)
	return fmt.Sprintf("%s %d, %d", t.Month(), t.Day(), t.Year())
}

func (p *PaymentDetails) PaymentAmountString() string {
	return fmt.Sprintf("$%d", p.PaymentAmount/100)
}

func (p *PaymentDetails) PlanAmountString() string {
	return fmt.Sprintf("$%d", p.PlanAmount/100)
}

func NewEmailTemplate(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User) *EmailTemplate {
	host := proxyServer.Host
	if admin.Host != "" {
		host = admin.Host
	}
	if _smtp.EmailHost != "" {
		host = _smtp.EmailHost
	}
	port := proxyServer.Port
	if _smtp.EmailPort != 0 {
		port = _smtp.EmailPort
	}
	return &EmailTemplate{
		From:   _smtp.Sender,
		To:     user.Email,
		Scheme: _smtp.EmailScheme,
		Host:   host,
		Port:   port,
		Prefix: admin.PathPrefix,
	}
}

func NewEmailTemplateWithPaymentDetails(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, paymentDetails *PaymentDetails) *EmailTemplate {
	emailTemplate := NewEmailTemplate(_smtp, proxyServer, admin, user)
	emailTemplate.PaymentDetails = paymentDetails
	return emailTemplate
}

func (e *EmailTemplate) UrlPrefix() string {
	if (e.Scheme == "http" && e.Port == 80) || (e.Scheme == "https" && e.Port == 443) {
		return fmt.Sprintf("%v://%v%v", e.Scheme, e.Host, e.Prefix)
	} else {
		return fmt.Sprintf("%v://%v:%v%v", e.Scheme, e.Host, e.Port, e.Prefix)
	}
}

func send(email string, _smtp config.SMTP, body []byte) error {
	auth := smtp.PlainAuth(
		"",
		_smtp.User,
		_smtp.Password,
		_smtp.Server,
	)
	err := smtp.SendMail(
		fmt.Sprintf("%v:%v", _smtp.Server, _smtp.Port),
		auth,
		_smtp.Sender,
		[]string{email},
		body,
	)
	if err != nil {
		logreport.Printf("error sending email %v", err)
	}
	return err
}

func Send(bodyTemplate string, data interface{}, _smtp config.SMTP, user *model.User, async bool) error {
	t := template.New("template")
	html, err := Asset("mail.html")
	if err != nil {
		return err
	}
	t, err = t.Parse(string(html))
	if err != nil {
		return err
	}
	html, err = Asset(bodyTemplate)
	if err != nil {
		return err
	}
	t, err = t.Parse(string(html))
	if err != nil {
		return err
	}
	var body bytes.Buffer
	err = t.Execute(&body, data)
	if err != nil {
		return err
	}

	if async {
		go send(user.Email, _smtp, body.Bytes())
	} else {
		return send(user.Email, _smtp, body.Bytes())
	}

	return nil
}
