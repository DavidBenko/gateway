package admin

import (
	"bytes"
	"fmt"
	"net/http"
	"net/smtp"
	"text/template"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

type PasswordResetController struct {
	BaseController
}

type PasswordReset struct {
	Email string `json:"email"`
}

func RoutePasswordReset(controller *PasswordResetController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"POST": write(db, controller.Reset),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"POST", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

type EmailTemplate struct {
	From  string
	To    string
	Host  string
	Port  int64
	Token string
}

var emailTemplate = `From: {{.From}}
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
  Click on the below link to reset your password:<br/>
  <a href="http://{{.Host}}:{{.Port}}/admin#/password/reset-confirmation?token={{.Token}}">reset password</a>
 </body>
</html>
`

func (c *PasswordResetController) Reset(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	request := struct {
		PasswordReset PasswordReset `json:"password_reset"`
	}{}
	if err := deserialize(&request, r.Body); err != nil {
		return err
	}

	email := request.PasswordReset.Email
	user, err := model.FindUserByEmail(tx.DB, email)
	if err != nil {
		return nil
	}
	_ = user

	token, err := model.AddUserToken(tx, email, model.TokenTypeReset)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	context := EmailTemplate{
		From:  c.SMTP.Sender,
		To:    email,
		Host:  c.ProxyServer.Host,
		Port:  c.ProxyServer.Port,
		Token: token,
	}
	t := template.New("template")
	t, err = t.Parse(emailTemplate)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	var body bytes.Buffer
	err = t.Execute(&body, context)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	auth := smtp.PlainAuth(
		"",
		c.SMTP.User,
		c.SMTP.Password,
		c.SMTP.Server,
	)
	err = smtp.SendMail(
		fmt.Sprintf("%v:%v", c.SMTP.Server, c.SMTP.Port),
		auth,
		c.SMTP.Sender,
		[]string{email},
		body.Bytes(),
	)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

type PasswordResetConfirmationController struct {
	BaseController
}

type PasswordResetConfirmation struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func RoutePasswordResetConfirmation(controller *PasswordResetConfirmationController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"POST": write(db, controller.Confirmation),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"POST", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

func (c *PasswordResetConfirmationController) Confirmation(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	request := struct {
		PasswordResetConfirmation PasswordResetConfirmation `json:"password_reset_confirmation"`
	}{}
	if err := deserialize(&request, r.Body); err != nil {
		return err
	}
	user, err := model.ValidateUserToken(tx, request.PasswordResetConfirmation.Token)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	user.NewPassword = request.PasswordResetConfirmation.NewPassword
	err = user.Update(tx)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
