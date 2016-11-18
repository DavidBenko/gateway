package model

import (
	"errors"
	aperrors "gateway/errors"
	apsql "gateway/sql"
)

type ProxyEndpointChannel struct {
	AccountID       int64 `json:"-"`
	UserID          int64 `json:"-"`
	APIID           int64 `json:"-" db:"api_id" path:"apiID"`
	ProxyEndpointID int64 `json:"proxy_endpoint_id,omitempty" db:"proxy_endpoint_id" path:"endpointID"`

	ID               int64  `json:"id,omitempty" path:"id"`
	RemoteEndpointID int64  `json:"remote_endpoint_id,omitempty" db:"remote_endpoint_id"`
	Name             string `json:"name"`

	// Export Indices
	ExportProxyEndpointIndex  int `json:"proxy_endpoint_index,omitempty"`
	ExportRemoteEndpointIndex int `json:"remote_endpoint_index,omitempty"`
}

func (c *ProxyEndpointChannel) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if c.RemoteEndpointID == 0 {
		errors.Add("remote_endpoint", "must be selected")
	}
	if c.Name == "" {
		errors.Add("name", "must have a name")
	}
	return errors
}

func (c *ProxyEndpointChannel) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "proxy_endpoint_channels", "remote_endpoint_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (c *ProxyEndpointChannel) All(db *apsql.DB) ([]*ProxyEndpointChannel, error) {
	channels := []*ProxyEndpointChannel{}
	var err error
	if c.APIID > 0 && c.AccountID > 0 {
		if c.ProxyEndpointID > 0 {
			err = db.Select(&channels, db.SQL("proxy_endpoint_channels/all"), c.ProxyEndpointID, c.APIID, c.AccountID)
		} else {
			err = db.Select(&channels, db.SQL("proxy_endpoint_channels/all_api"), c.APIID, c.AccountID)
		}
	} else {
		err = errors.New("APIID and AccountID required for All")
	}
	if err != nil {
		return nil, err
	}
	for _, channel := range channels {
		channel.AccountID = c.AccountID
		channel.UserID = c.UserID
	}
	return channels, nil
}

func (c *ProxyEndpointChannel) Find(db *apsql.DB) (*ProxyEndpointChannel, error) {
	channel := ProxyEndpointChannel{
		AccountID: c.AccountID,
		UserID:    c.UserID,
	}
	err := db.Get(&channel, db.SQL("proxy_endpoint_channels/find"), c.ID, c.ProxyEndpointID, c.APIID, c.AccountID)
	return &channel, err
}

func (c *ProxyEndpointChannel) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("proxy_endpoint_channels/delete"), c.ID, c.ProxyEndpointID, c.APIID, c.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("proxy_endpoint_channels", c.AccountID, c.UserID, c.APIID, c.ProxyEndpointID, c.ID, apsql.Delete)
}

func (c *ProxyEndpointChannel) Insert(tx *apsql.Tx) error {
	var err error
	c.ID, err = tx.InsertOne(tx.SQL("proxy_endpoint_channels/insert"), c.ProxyEndpointID,
		c.APIID, c.AccountID, c.RemoteEndpointID, c.APIID, c.AccountID, c.Name)
	if err != nil {
		return err
	}
	return tx.Notify("proxy_endpoint_channels", c.AccountID, c.UserID, c.APIID, c.ProxyEndpointID, c.ID, apsql.Insert)
}

func (c *ProxyEndpointChannel) Update(tx *apsql.Tx) error {
	err := tx.UpdateOne(tx.SQL("proxy_endpoint_channels/update"),
		c.RemoteEndpointID, c.APIID, c.AccountID, c.Name, c.ID,
		c.ProxyEndpointID, c.APIID, c.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("proxy_endpoint_channels", c.AccountID, c.UserID, c.APIID, c.ProxyEndpointID, c.ID, apsql.Update)
}
