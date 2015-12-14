package mail

import (
	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

const confirmTemplate = `{{define "body"}}
  Click on the below link to confirm your email:<br/>
  <a href="{{.UrlPrefix}}confirmation?token={{.Token}}">confirm email</a>
{{end}}
`

func SendConfirmEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, tx *apsql.Tx) error {
	token, err := model.AddUserToken(tx, user.Email, model.TokenTypeConfirm)
	if err != nil {
		return err
	}

	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "JustAPIs Email Confirmation"
	context.Token = token
	err = Send(confirmTemplate, context, _smtp, user)
	if err != nil {
		return err
	}

	return nil
}
