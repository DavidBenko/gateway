package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"gateway/config"
	"gateway/logreport"
	"gateway/model"
)

type EmailTemplate struct {
	From    string
	To      string
	Subject string
	Scheme  string
	Host    string
	Port    int64
	Prefix  string
	Token   string
	Resend  bool
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
