package mail

import (
	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

const resetTemplate = `{{define "body"}}
  Click on the below link to reset your password:<br/>
  <a href="{{.UrlPrefix}}password_reset_check?token={{.Token}}">reset password</a>
{{end}}
`

func SendResetEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, tx *apsql.Tx) error {
	token, err := model.AddUserToken(tx, user.Email, model.TokenTypeReset)
	if err != nil {
		return err
	}

	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "JustAPIs Password Reset"
	context.Token = token
	err = Send(resetTemplate, context, _smtp, user)
	if err != nil {
		return err
	}

	return nil
}
