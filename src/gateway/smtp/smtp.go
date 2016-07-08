package smtp

import (
	"encoding/json"
	"fmt"
	"gateway/logreport"
	"net/smtp"
)

type Mailer interface {
	Send(email string, body string) error
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

func (s *Spec) Send(email string, body string) error {
	err := smtp.SendMail(
		fmt.Sprintf("%v:%v", s.Host, s.Port),
		s.Auth,
		s.Sender,
		[]string{email},
		[]byte(body),
	)

	if err != nil {
		logreport.Printf("error sending email using %v:%v : %v", s.Host, s.Port, err)
	}

	return err
}

func (s *Spec) ConnectionString() string {
	spec, err := json.Marshal(s)

	if err != nil {
		logreport.Fatal(err)
	}

	return string(spec)
}
