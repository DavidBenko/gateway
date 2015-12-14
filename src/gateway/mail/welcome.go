package mail

import (
	"gateway/config"
	"gateway/model"
)

const welcomeTemplate = `{{define "body"}}
  Welcome to JustAPIs!<br/>
  To view the docs click <a href="http://docs.justapis.com/">here</a><br/>
  To login click <a href="{{.UrlPrefix}}#/login">here</a>
{{end}}
`

func SendWelcomeEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User) error {
	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "Welcome to JustAPIs!"
	err := Send(welcomeTemplate, context, _smtp, user)
	if err != nil {
		return err
	}

	return nil
}
