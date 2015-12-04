package admin

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

type sanitizedUser struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Admin     bool   `json:"admin"`
	Confirmed bool   `json:"confirmed"`
}

func (c *UsersController) sanitize(user *model.User) *sanitizedUser {
	return &sanitizedUser{user.ID, user.Name, user.Email, user.Admin, user.Confirmed}
}

var confirmTemplate = `From: {{.From}}
To: {{.To}}
Subject: JustAPIs Password Reset
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";

<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
	<title>JustAPIs Password Reset</title>
	<style type="text/css">
	 body{width:100% !important; -webkit-text-size-adjust:100%; -ms-text-size-adjust:100%; margin:0; padding:0;}
	</style>
 </head>
 <body>
  Click on the below link to confirm you email:<br/>
  <a href="http://{{.Host}}:{{.Port}}/admin/confirmation?token={{.Token}}">confirm email</a>
 </body>
</html>
`

func SendConfirmEmail(_smtp config.SMTP, proxyServer config.ProxyServer, user *model.User, tx *apsql.Tx) error {
	token, err := model.AddUserToken(tx, user.Email, model.TokenTypeConfirm)
	if err != nil {
		return err
	}

	context := EmailTemplate{
		From:  _smtp.Sender,
		To:    user.Email,
		Host:  proxyServer.Host,
		Port:  proxyServer.Port,
		Token: token,
	}
	t := template.New("template")
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

func (c *UsersController) AfterInsert(user *model.User, tx *apsql.Tx) error {
	if user.Confirmed {
		return nil
	}

	return SendConfirmEmail(c.SMTP, c.ProxyServer, user, tx)
}
