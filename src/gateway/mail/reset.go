package mail

import (
	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

const resetTemplate = `{{define "body"}}
  <p>
	  Click on the link below to reset your password:<br/>
		<a href="{{.UrlPrefix}}password_reset_check?token={{.Token}}">reset password</a>
	</p>
  <p>
	  If you have any questions or require support, please contact
		<a href="mailto:support@anypresence.com">support@anypresence.com</a><br/>
		- The AnyPresence Team
	</p>
{{end}}
`

func SendResetEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, tx *apsql.Tx, async bool) error {
	token, err := model.AddUserToken(tx, user.Email, model.TokenTypeReset)
	if err != nil {
		return err
	}

	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "JustAPIs Password Reset"
	context.Token = token
	err = Send(resetTemplate, context, _smtp, user, async)
	if err != nil {
		return err
	}

	return nil
}
