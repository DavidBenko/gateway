package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/logreport"
	"gateway/model"
	"gateway/smtp"
)

type SmtpRequest struct {
	Config  *smtp.Spec `json:"config"`
	Body    string     `json:"body"`
	Address string     `json:"address"`
	mailer  smtp.Mailer
}

type SmtpResponse struct {
	Data map[string]interface{} `json:"data"`
}

func NewSmtpRequest(pool *smtp.SmtpPool, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &SmtpRequest{}

	if pool == nil {
		return nil, errors.New("smtp pool not setup")
	}

	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	endpointData := &SmtpRequest{}
	if err := json.Unmarshal(endpoint.Data, endpointData); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal endpoint configuration: %v", err)
	}

	request.updateWith(endpointData)

	mailer, err := pool.Connection(request.Config)

	if err != nil {
		return nil, err
	}

	request.mailer = mailer

	return request, nil
}

func (r *SmtpRequest) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *SmtpRequest) Log(devMode bool) string {
	s := fmt.Sprintf("To: %s\n", r.Address)

	if devMode {
		s += fmt.Sprintf("Body: %s\nConfiguration: %+v\n", r.Body, r.Config)
	}

	return s
}

func (r *SmtpRequest) Perform() Response {
	response := &SmtpResponse{}
	data := make(map[string]interface{}, 1)

	data["success"] = true

	err := r.mailer.Send(r.Address, r.Body)

	if err != nil {
		logreport.Print(err)
		data["success"] = false
	}

	response.Data = data

	return response
}

func (r *SmtpRequest) updateWith(endpiontData *SmtpRequest) {
	if endpiontData.Config != nil {
		r.Config = endpiontData.Config
	} else {
		if r.Config == nil {
			r.Config = &smtp.Spec{}
		}
	}
}

func (r *SmtpResponse) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *SmtpResponse) Log() string {
	return fmt.Sprintf("Success: %t", r.Data["success"])
}
