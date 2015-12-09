package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

const confirmTemplate = `{{define "body"}}
  Click on the below link to confirm your email:<br/>
  <a href="http://{{.Host}}:{{.Port}}{{.Prefix}}confirmation?token={{.Token}}">confirm email</a>
{{end}}
`

func SendConfirmEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, tx *apsql.Tx) error {
	token, err := model.AddUserToken(tx, user.Email, model.TokenTypeConfirm)
	if err != nil {
		return err
	}

	host := proxyServer.Host
	if admin.Host != "" {
		host = admin.Host
	}
	context := EmailTemplate{
		From:    _smtp.Sender,
		To:      user.Email,
		Subject: "JustAPIs Email Confirmation",
		Host:    host,
		Port:    proxyServer.Port,
		Prefix:  admin.PathPrefix,
		Token:   token,
	}
	t := template.New("template")
	t, err = t.Parse(mailTemplate)
	if err != nil {
		return err
	}
	t, err = t.Parse(confirmTemplate)
	if err != nil {
		return err
	}
	var body bytes.Buffer
	err = t.Execute(&body, context)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth(
		"",
		_smtp.User,
		_smtp.Password,
		_smtp.Server,
	)
	err = smtp.SendMail(
		fmt.Sprintf("%v:%v", _smtp.Server, _smtp.Port),
		auth,
		_smtp.Sender,
		[]string{user.Email},
		body.Bytes(),
	)
	if err != nil {
		return err
	}

	return nil
}
