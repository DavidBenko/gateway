package remote_endpoint

import (
	"gateway/docker"
	aperrors "gateway/errors"
)

// Docker represents a configuration for a remote Docker endpoint
type Docker struct {
	Repository  string            `json:"repository"`
	Tag         string            `json:"tag"`
	Command     string            `json:"command"`
	Arguments   []string          `json:"arguments"`
	Environment map[string]string `json:"environment"`
	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	Registry    string            `json:"registry,omitempty"`
}

// Validate validates the existence of an image and a command
func (d *Docker) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)

	if d.Repository == "" {
		errors.Add("repository", "must not be blank")
	}

	if d.Tag == "" {
		errors.Add("tag", "must not be blank")
	}

	if d.Password != "" && d.Username == "" {
		errors.Add("username", "must not be blank when a password is provided")
	}

	if d.Username != "" && d.Password == "" {
		errors.Add("password", "must not be blank when a username is provided")
	}

	// No blank arguments
	if len(d.Arguments) > 0 {
		for _, v := range d.Arguments {
			if v == "" {
				errors.Add("arguments", "blank arguments are prohibited")
				return errors
			}
		}
	}

	// Make sure we can access the image.
	if d.Repository != "" && d.Tag != "" {
		dc := &docker.DockerConfig{Repository: d.Repository, Tag: d.Tag, Username: d.Username, Password: d.Password, Registry: d.Registry}
		exists, err := dc.ImageExists()
		if err != nil {
			errors.Add("repository", err.Error())
		}
		if !exists {
			errors.Add("repository", "could not find image in registry")
		}
	}

	if !errors.Empty() {
		return errors
	}

	return nil
}
