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

const resetTemplate = `{{define "body"}}
  Click on the below link to reset your password:<br/>
  <a href="{{.UrlPrefix}}#/password/reset-confirmation?token={{.Token}}">reset password</a>
{{end}}
`

func SendResetEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, tx *apsql.Tx) error {
	token, err := model.AddUserToken(tx, user.Email, model.TokenTypeReset)
	if err != nil {
		return err
	}

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
	context := &EmailTemplate{
		From:    _smtp.Sender,
		To:      user.Email,
		Subject: "JustAPIs Password Reset",
		Scheme:  _smtp.EmailScheme,
		Host:    host,
		Port:    port,
		Prefix:  admin.PathPrefix,
		Token:   token,
	}
	t := template.New("template")
	t, err = t.Parse(mailTemplate)
	if err != nil {
		return err
	}
	t, err = t.Parse(resetTemplate)
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
