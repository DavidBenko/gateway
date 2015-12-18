package mail

import (
	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

const confirmTemplate = `{{define "body"}}
  <p>Thanks for your interest in JustAPIs.</p>
	<p>
	  Click on the link below to confirm your email:<br/>
    <a href="{{.UrlPrefix}}confirmation?token={{.Token}}">confirm email</a>
	</p>
	<p>
	  If you have any questions or require support, please contact
	  <a href="mailto:support@anypresence.com">support@anypresence.com</a>
	</p>
	<p>
	  Thank you<br/>
    -The AnyPresence Team
	</p>
{{end}}
`

func SendConfirmEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, tx *apsql.Tx, async bool) error {
	token, err := model.AddUserToken(tx, user.Email, model.TokenTypeConfirm)
	if err != nil {
		return err
	}

	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "JustAPIs Email Confirmation"
	context.Token = token
	err = Send(confirmTemplate, context, _smtp, user, async)
	if err != nil {
		return err
	}

	return nil
}
