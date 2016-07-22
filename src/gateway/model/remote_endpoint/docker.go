package remote_endpoint

import aperrors "gateway/errors"

// Docker represents a configuration for a remote Docker endpoint
type Docker struct {
	Image     string   `json:"image"`
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
}

// Validate validates the existence of an image and a command
func (d *Docker) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)

	if d.Image == "" {
		errors.Add("image", "must not be blank")
	}

	if d.Command == "" {
		errors.Add("command", "must not be blank")
	}

	if len(d.Arguments) > 0 {
		for i := range d.Arguments {
			if d.Arguments[i] == "" {
				errors.Add("arguments", "must not contain blank arguments")
			}
		}
	}

	if !errors.Empty() {
		return errors
	}

	return nil
}
