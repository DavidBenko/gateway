package remote_endpoint

import (
	aperrors "gateway/errors"
	"gateway/smtp"
)

type Smtp struct {
	Config *smtp.Spec
}

func (s *Smtp) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)

	if s.Config.Username == "" {
		errors.Add("username", "username is required")
	}

	if s.Config.Password == "" {
		errors.Add("password", "password is required")
	}

	if s.Config.Host == "" {
		errors.Add("host", "host is required")
	}

	if s.Config.Port == 0 {
		errors.Add("port", "port is required")
	}

	if s.Config.Sender == "" {
		errors.Add("sender", "sender is required")
	}

	return errors
}
