package mail

import (
	"gateway/config"
	"gateway/model"
)

const welcomeTemplate = `{{define "body"}}
  {{if .Resend}}
	<p>
	  <b>
		  You tried to register an already registered and confirmed email address.
		  If you forgot your password, please visit the password reset link to get a new one.
		</b>
	</p>
	{{end}}
	<p>Thanks for confirming your JustAPIs account email.</p>
	<p><b>To use the online version of JustAPIs, click <a href="{{.UrlPrefix}}#/login">here</a></b></p>
	<p style="margin-bottom: 0px;"><b>To download JustAPIs, select the appropriate install package (zip archive):</b></p>
	<ul style="margin-top: 0px; margin-bottom: 0px; padding-left: 10px;">
	  <li><a href="http://www2.anypresence.com/mac">Mac OSX</a></li>
		<li><a href="http://www2.anypresence.com/linux64">Linux 64-bit</a></li>
		<li><a href="http://www2.anypresence.com/linux32">Linux 32-bit</a></li>
		<li><a href="http://www2.anypresence.com/win32">Windows 32-bit</a></li>
		<li><a href="http://www2.anypresence.com/win64">Windows 64-bit</a></li>
		<li><a href="http://www2.anypresence.com/linux_armv6">Linux ARMv6</a></li>
		<li><a href="http://www2.anypresence.com/linux_armv7">Linux ARMv7</a></li>
	</ul>
	<p style="margin-top: 0px;">
		<b>
		  To learn how to use JustAPIs, follow instructions in this
			<a href="http://docs.justapis.com/quickstart">Quick Start Guide</a>
		</b>
	</p>
	<p>
    If you have any questions or require support, please contact
		<a href="mailto:support@anypresence.com">support@anypresence.com</a><br/>
		- The AnyPresence Team
	</p>
	<p>P.S. Don't forget to take the <a href="http://tshirt.justapis.com/">JustAPIs T-Shirt Challenge!</a></p>
{{end}}
`

func SendWelcomeEmail(_smtp config.SMTP, proxyServer config.ProxyServer, admin config.ProxyAdmin,
	user *model.User, resend, async bool) error {
	context := NewEmailTemplate(_smtp, proxyServer, admin, user)
	context.Subject = "Welcome to JustAPIs!"
	context.Resend = resend
	err := Send(welcomeTemplate, context, _smtp, user, async)
	if err != nil {
		return err
	}

	return nil
}
