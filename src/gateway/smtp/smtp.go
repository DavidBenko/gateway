package smtp

import (
	"encoding/json"
	"errors"
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
	User     string `json:"user"`
	Password string `json:"password"`
	Sender   string `json:"sender"`
	Auth     smtp.Auth
}

func NewSpec(host string, port int, user string, password string, sender string) (*Spec, error) {
	if user == "" {
		return nil, errors.New("user is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	auth := smtp.PlainAuth(
		"",
		user,
		password,
		host,
	)

	return &Spec{host, port, user, password, sender, auth}, nil
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
