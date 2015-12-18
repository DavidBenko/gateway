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

const mailTemplate = `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";

<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
	<title>{{.Subject}}</title>
	<style type="text/css">
	 body{width:100% !important; -webkit-text-size-adjust:100%; -ms-text-size-adjust:100%; margin:0; padding:0;}
	 ul li {
    color: #386EFF;
	 }
	 ul li a {
    color: blue;
	 }
	</style>
 </head>
 <body>{{template "body" .}}</body>
</html>
`

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

func Send(text string, data interface{}, _smtp config.SMTP, user *model.User, async bool) error {
	t := template.New("template")
	t, err := t.Parse(mailTemplate)
	if err != nil {
		return err
	}
	t, err = t.Parse(text)
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
