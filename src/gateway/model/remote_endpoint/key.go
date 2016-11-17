package remote_endpoint

import (
	aperrors "gateway/errors"
	"gateway/key"
	"strings"
)

type Key struct {
	Config *key.Spec `json:"config"`
}

func (s *Key) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)

	if strings.TrimSpace(s.Config.Name) == "" {
		errors.Add("name", "name is required")
	}

	return errors
}
