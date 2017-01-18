package request

import (
	"bytes"
	"encoding/json"
	"fmt"

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

	runOutput, err := model.ExecuteCustomFunction(r.db, r.AccountID, r.APIID, 0, name, input)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	lines, output := runOutput.Parts()

	err = json.Unmarshal([]byte(output), &response.Output)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.LogLines = lines
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
	LogLines   []string               `json:"log_lines"`
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

func (r *CustomFunctionResponse) Logs() []string {
	return r.LogLines
}

func NewCustomFunctionRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage, db *sql.DB) (Request, error) {
	request := &CustomFunctionRequest{AccountID: endpoint.AccountID, APIID: endpoint.APIID, db: db}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	return request, nil
}
