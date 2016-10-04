package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"gateway/config"
	"gateway/http"
	"gateway/model"
	sql "gateway/sql"
)

type ExecuteJob func(jobID, accountID, apiID int64, logPrefix, attributes string) (err error)

type JobRequest struct {
	Arguments  map[string]interface{} `json:"arguments"`
	DB         *sql.DB                `json:"-"`
	AccountID  int64                  `json:"-"`
	APIID      int64                  `json:"-"`
	ExecuteJob ExecuteJob             `json:"-"`
}

type jobOperation func(request *JobRequest) error

func storeOperationRun(request *JobRequest) error {
	name, valid := request.Arguments["1"].(string)
	if !valid {
		return errors.New("name is not a string")
	}

	attributes, valid := request.Arguments["2"].(interface{})
	if !valid {
		return errors.New("attributes is not an Object")
	}

	endpoint := &model.ProxyEndpoint{
		Name:      name,
		Type:      model.ProxyEndpointTypeJob,
		APIID:     request.APIID,
		AccountID: request.AccountID,
	}
	endpoint, err := endpoint.Find(request.DB)
	if err != nil {
		return err
	}

	logPrefix := fmt.Sprintf("%s [act %d] [api %d] [end %d]", config.Job,
		endpoint.AccountID, endpoint.APIID, endpoint.ID)

	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return err
	}

	go request.ExecuteJob(endpoint.ID, endpoint.AccountID, endpoint.APIID, logPrefix, string(attributesJSON))

	return nil
}

func storeOperationSchedule(request *JobRequest) error {
	time, valid := request.Arguments["1"].(float64)
	if !valid {
		return errors.New("time is not a number")
	}

	name, valid := request.Arguments["2"].(string)
	if !valid {
		return errors.New("name is not a string")
	}

	attributes, valid := request.Arguments["3"].(interface{})
	if !valid {
		return errors.New("attributes is not an Object")
	}

	endpoint := &model.ProxyEndpoint{
		Name:      name,
		Type:      model.ProxyEndpointTypeJob,
		APIID:     request.APIID,
		AccountID: request.AccountID,
	}
	endpoint, err := endpoint.Find(request.DB)
	if err != nil {
		return err
	}

	uuid, err := http.NewUUID()
	if err != nil {
		return err
	}

	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return err
	}

	timer := &model.Timer{
		AccountID:  request.AccountID,
		APIID:      request.APIID,
		JobID:      endpoint.ID,
		Name:       uuid,
		Once:       true,
		Next:       int64(time),
		Attributes: attributesJSON,
	}
	err = request.DB.DoInTransaction(func(tx *sql.Tx) error {
		return timer.Insert(tx)
	})
	if err != nil {
		return err
	}

	return nil
}

var jobOperations = map[string]jobOperation{
	"run":      storeOperationRun,
	"schedule": storeOperationSchedule,
}

func (r *JobRequest) Perform() (_response Response) {
	response := &JobResponse{}
	_response = response
	defer func() {
		if r := recover(); r != nil {
			response.Error = fmt.Sprintf("%v", r)
		}
	}()

	op := r.Arguments["0"]
	if _, valid := op.(string); !valid {
		response.Error = "Missing operation parameter"
		return
	}
	if op, valid := jobOperations[op.(string)]; valid {
		err := op(r)
		if err != nil {
			response.Error = err.Error()
		}
	} else {
		response.Error = "Invalid operation"
	}

	return
}

func (r *JobRequest) Log(devMode bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("\nArguments: %s", r.Arguments))
	if devMode {
		buffer.WriteString(fmt.Sprintf("\nAccountID: %d APIID: %d", r.AccountID, r.APIID))
	}
	return buffer.String()
}

func (r *JobRequest) JSON() ([]byte, error) {
	return json.Marshal(r)
}

type JobResponse struct {
	Error string `json:"error,omitempty"`
}

func (r *JobResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

func (r *JobResponse) Log() string {
	return r.Error
}

func NewJobRequest(db *sql.DB, endpoint *model.RemoteEndpoint, executeJob ExecuteJob, data *json.RawMessage) (Request, error) {
	request := &JobRequest{
		DB:         db,
		AccountID:  endpoint.AccountID,
		APIID:      endpoint.APIID,
		ExecuteJob: executeJob,
	}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	return request, nil
}
