package admin

import (
	"encoding/json"
	"fmt"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

// BeforeInsert does some work before inserting a RemoteEndpoint
func (c *RemoteEndpointsController) BeforeInsert(remoteEndpoint *model.RemoteEndpoint, tx *apsql.Tx) error {
	if remoteEndpoint.Type != model.RemoteEndpointTypeSoap {
		return nil
	}

	soap, err := model.NewSoapRemoteEndpoint(remoteEndpoint)
	if err != nil {
		return fmt.Errorf("Unable to construct SoapRemoteEndpoint object: %v", err)
	}

	remoteEndpoint.Soap = &soap
	remoteEndpoint.Data = types.JsonText(json.RawMessage([]byte("null")))

	return nil
}

// BeforeUpdate does some work before updataing a RemoteEndpoint
func (c *RemoteEndpointsController) BeforeUpdate(remoteEndpoint *model.RemoteEndpoint, tx *apsql.Tx) error {
	if remoteEndpoint.Type != model.RemoteEndpointTypeSoap {
		return nil
	}

	// TODO

	return nil
}

// AfterInsert does some work after inserting a RemoteEndpoint
func (c *RemoteEndpointsController) AfterInsert(remoteEndpoint *model.RemoteEndpoint, tx *apsql.Tx) error {
	if remoteEndpoint.Type != model.RemoteEndpointTypeSoap {
		return nil
	}

	remoteEndpoint.Soap.RemoteEndpointID = remoteEndpoint.ID
	err := remoteEndpoint.Soap.Insert(tx)
	if err != nil {
		return fmt.Errorf("Unable to insert SoapRemoteEndpoint: %v: ", err)
	}

	return nil
}

// AfterUpdate does some work after updating a RemoteEndpoint
func (c *RemoteEndpointsController) AfterUpdate(remoteEndpoint *model.RemoteEndpoint, tx *apsql.Tx) error {
	if remoteEndpoint.Type != model.RemoteEndpointTypeSoap {
		return nil
	}

	// TODO

	return nil
}
