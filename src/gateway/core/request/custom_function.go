package request

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gateway/docker"
	"gateway/model"
	"gateway/sql"
)

type CustomFunctionRequest struct {
	Arguments map[string]interface{} `json:"arguments"`
	AccountID int64                  `json:"-"`
	APIID     int64                  `json:"-"`
	db        *sql.DB
}

func (r *CustomFunctionRequest) Perform() Response {
	response := &CustomFunctionResponse{}

	op := r.Arguments["0"]
	if op, valid := op.(string); !valid || op != "call" {
		response.Error = "Invalid operation"
		return response
	}

	name, valid := r.Arguments["1"].(string)
	if !valid {
		response.Error = "name is not a string"
		return response
	}

	input, valid := r.Arguments["2"].(interface{})
	if !valid {
		response.Error = "input is not an object"
		return response
	}

	function := &model.CustomFunction{
		AccountID: r.AccountID,
		APIID:     r.APIID,
		Name:      name,
	}
	function, err := function.Find(r.db)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	if !function.Active {
		response.Error = "Custom function is not active"
		return response
	}

	runOutput, err := docker.ExecuteImage(function.ImageName(), input)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	err = json.Unmarshal([]byte(runOutput.Stdout), &response.Output)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Stderr = runOutput.Stderr
	response.Logs = runOutput.Logs
	response.StatusCode = runOutput.StatusCode

	return response
}

func (r *CustomFunctionRequest) Log(devMode bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("\nArguments: %s", r.Arguments))
	if devMode {
		buffer.WriteString(fmt.Sprintf("\nAccountID: %d\nAPIID: %d", r.AccountID, r.APIID))
	}
	return buffer.String()
}

func (r *CustomFunctionRequest) JSON() ([]byte, error) {
	return json.Marshal(r)
}

type CustomFunctionResponse struct {
	Output     map[string]interface{} `json:"output"`
	Stderr     string                 `json:"stderr"`
	Logs       string                 `json:"logs"`
	StatusCode int                    `json:"status_code"`
	Error      string                 `json:"error,omitempty"`
}

func (r *CustomFunctionResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

func (r *CustomFunctionResponse) Log() string {
	if r.Output != nil {
		return "Custom function successful"
	}

	return r.Error
}

func NewCustomFunctionRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage, db *sql.DB) (Request, error) {
	request := &CustomFunctionRequest{AccountID: endpoint.AccountID, APIID: endpoint.APIID, db: db}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	return request, nil
}
