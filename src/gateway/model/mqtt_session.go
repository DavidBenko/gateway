package model

import (
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type MQTTSession struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"api_id" db:"api_id"`

	ID               int64          `json:"id,omitempty"`
	RemoteEndpointID int64          `json:"remote_endpoint_id" db:"remote_endpoint_id"`
	Type             string         `json:"type"`
	ClientID         string         `json:"client_id" db:"client_id"`
	Data             types.JsonText `json:"data"`
}

func (s *MQTTSession) All(db *apsql.DB) ([]*MQTTSession, error) {
	sessions := []*MQTTSession{}
	err := db.Select(&sessions, db.SQL("mqtt_sessions/all"), s.APIID, s.AccountID)
	for _, session := range sessions {
		session.AccountID = s.AccountID
		session.APIID = s.APIID
	}
	return sessions, err
}

func (s *MQTTSession) Count(db *apsql.DB) int {
	var count int
	db.Get(&count, db.SQL("mqtt_sessions/count"), s.RemoteEndpointID,
		s.Type, s.APIID, s.AccountID)
	return count
}

func (s *MQTTSession) Find(db *apsql.DB) (*MQTTSession, error) {
	session := MQTTSession{
		AccountID: s.AccountID,
		APIID:     s.APIID,
	}
	err := db.Get(&session, db.SQL("mqtt_sessions/find"), s.RemoteEndpointID,
		s.Type, s.ClientID, s.APIID, s.AccountID)
	return &session, err
}

func (s *MQTTSession) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("mqtt_sessions/delete"), s.Type, s.ClientID,
		s.RemoteEndpointID, s.APIID, s.AccountID)
	return err
}

func (s *MQTTSession) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}

	_, err = tx.InsertOne(tx.SQL("mqtt_sessions/insert"), s.RemoteEndpointID,
		s.APIID, s.AccountID, s.Type, s.ClientID, data)
	return err
}

func (s *MQTTSession) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("mqtt_sessions/update"), data, s.Type, s.ClientID,
		s.RemoteEndpointID, s.APIID, s.AccountID)
	return err
}
