package request

import (
	"encoding/json"
	"fmt"

	"gateway/model"
	re "gateway/model/remote_endpoint"
	"gateway/push"
	sql "gateway/sql"
)

type PushRequest struct {
	re.Push
	Channel          string                 `json:"channel"`
	Payload          map[string]interface{} `json:"payload"`
	pool             *push.PushPool
	db               *sql.DB
	accountID        int64
	apiID            int64
	remoteEndpointID int64
}

type PushResponse struct {
	Error string `json:"error"`
}

func NewPushRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage, pool *push.PushPool, db *sql.DB) (Request, error) {
	request := &PushRequest{
		pool:             pool,
		db:               db,
		accountID:        endpoint.AccountID,
		apiID:            endpoint.APIID,
		remoteEndpointID: endpoint.ID,
	}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, err
	}

	push := &re.Push{}
	err := json.Unmarshal(endpoint.Data, push)
	if err != nil {
		return nil, err
	}
	if endpoint.SelectedEnvironmentData != nil {
		if err := json.Unmarshal(*endpoint.SelectedEnvironmentData, request); err != nil {
			return nil, err
		}
	}
	request.UpdateWith(push)

	return request, nil
}

func (p *PushRequest) Perform() Response {
	response := &PushResponse{}
	err := p.db.DoInTransaction(func(tx *sql.Tx) error {
		return p.pool.Push(&p.Push, tx, p.accountID, p.apiID, p.remoteEndpointID, p.Channel, p.Payload)
	})
	if err != nil {
		response.Error = err.Error()
	}

	return response
}

func (p *PushRequest) Log(devMode bool) string {
	return fmt.Sprintf("%v %v", p.Channel, p.Payload)
}

func (p *PushRequest) JSON() ([]byte, error) {
	return json.Marshal(&p)
}

func (p *PushResponse) JSON() ([]byte, error) {
	return json.Marshal(&p)
}

func (p *PushResponse) Log() string {
	return p.Error
}
