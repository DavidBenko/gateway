package smtp

import (
	"encoding/json"
	"fmt"
	"gateway/logreport"
	"net/smtp"
	"strings"
)

type Mailer interface {
	Send(to, cc, bcc []string, body, subject string, html bool) error
	ConnectionString() string
}

type Spec struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Sender   string `json:"sender"`
	Auth     smtp.Auth
}

func (s *Spec) CreateAuth() {
	auth := smtp.PlainAuth(
		"",
		s.Username,
		s.Password,
		s.Host,
	)

	s.Auth = auth
}

func (s *Spec) Send(to, cc, bcc []string, body, subject string, html bool) error {
	allRecipients := append(append(to, bcc...), cc...)
	err := smtp.SendMail(
		fmt.Sprintf("%v:%v", s.Host, s.Port),
		s.Auth,
		s.Sender,
		allRecipients,
		[]byte(s.createBody(to, cc, body, subject, html)),
	)

	if err != nil {
		logreport.Printf("\nerror sending email using %v:%v : %v", s.Host, s.Port, err)
	}

	return err
}

func (s *Spec) createBody(addresses, cc []string, body, subject string, html bool) string {
	b := fmt.Sprintf("From: %s\r\n", s.Sender)
	b += fmt.Sprintf("To: %s\r\n", strings.Join(addresses[:], ","))
	if len(cc) > 0 {
		b += fmt.Sprintf("Cc: %s\r\n", strings.Join(cc[:], ","))
	}
	if html {
		b += "MIME-Version: 1.0\r\n"
		b += "Content-type: text/html\r\n"
	}
	b += fmt.Sprintf("Subject: %s\r\n\r\n", subject)
	b += fmt.Sprintf("%s\r\n", body)
	return b
}

func (s *Spec) ConnectionString() string {
	spec, err := json.Marshal(s)

	if err != nil {
		logreport.Fatal(err)
	}

	return string(spec)
}
