package model

import "encoding/json"

// Errors represents API-serializable validation errors.
type Errors map[string][]string

func (e Errors) add(name, message string) {
	e[name] = append(e[name], message)
}

// Empty reports if there are no errors.
func (e Errors) Empty() bool {
	return len(e) == 0
}

// JSON returns the errors' JSON representation for the API.
func (e Errors) JSON() ([]byte, error) {
	wrapped := struct {
		Errors Errors `json:"errors"`
	}{e}
	return json.MarshalIndent(wrapped, "", "    ")
}
