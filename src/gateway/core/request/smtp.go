package request

import (
	"encoding/json"
	"fmt"
	"gateway/model"
	"gateway/smtp"
)

type SmtpRequest struct {
	Config *smtp.Spec `json:"config"`
	Body   string     `json:"body"`
	Target string     `json:"target"`
	mailer smtp.Mailer
}

type SmtpResponse struct {
	Success bool
}

func NewSmtpRequest(pool *smtp.SmtpPool, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &SmtpRequest{}

	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to marshal request json: %v", err)
	}

	return request, nil
}

func (r *SmtpRequest) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *SmtpRequest) Log(devMode bool) string {
	s := fmt.Sprintf("To: %s\nBody: %s\n", r.Target, r.Body)

	if devMode {
		s += fmt.Sprintf("Configuration: %+v\n", r.Config)
	}

	return s
}

func (r *SmtpRequest) Perform() Response {
	response := &SmtpResponse{}

	return response
}

func (r *SmtpResponse) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *SmtpResponse) Log() string {
	return fmt.Sprintf("Success: %t", r.Success)
}
