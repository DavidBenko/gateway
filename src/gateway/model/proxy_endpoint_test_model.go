package model

import (
	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

const (
	ProxyEndpointTestMethodGet    = "GET"
	ProxyEndpointTestMethodPost   = "POST"
	ProxyEndpointTestMethodPut    = "PUT"
	ProxyEndpointTestMethodDelete = "DELETE"
)

type ProxyEndpointTest struct {
	ID        int64                    `json:"id,omitempty"`
	Name      string                   `json:"name"`
	Channels  bool                     `json:"channels"`
	ChannelID *int64                   `json:"channel_id" db:"channel_id"`
	Methods   types.JsonText           `json:"methods"`
	Route     string                   `json:"route"`
	Body      string                   `json:"body"`
	Pairs     []*ProxyEndpointTestPair `json:"pairs,omitempty"`
	Data      types.JsonText           `json:"data,omitempty"`
}

func (t *ProxyEndpointTest) GetMethods() (methods []string, err error) {
	err = t.Methods.Unmarshal(&methods)
	return
}

func (t *ProxyEndpointTest) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)
	if t.Name == "" {
		errors.Add("name", "must not be blank")
	}

	methods, err := t.GetMethods()
	if err != nil {
		errors.Add("methods", "must be valid json")
	} else if len(methods) == 0 {
		errors.Add("methods", "must be selected")
	} else {
		for _, method := range methods {
			switch method {
			case ProxyEndpointTestMethodGet:
			case ProxyEndpointTestMethodPost:
			case ProxyEndpointTestMethodPut:
			case ProxyEndpointTestMethodDelete:
			default:
				errors.Add("methods", "invalid http method")
			}
		}
	}

	if t.Channels {
		if t.ChannelID == nil {
			errors.Add("channel", "must be selected")
		}
	} else {
		if t.Route == "" {
			errors.Add("route", "must not be empty")
		}
	}
	return errors
}

func (t *ProxyEndpointTest) Insert(tx *apsql.Tx, endpointID int64) error {
	data, err := marshaledForStorage(t.Data)
	if err != nil {
		return aperrors.NewWrapped("Marshaling test JSON", err)
	}

	methods, err := marshaledForStorage(t.Methods)
	if err != nil {
		return aperrors.NewWrapped("Marshaling test methods JSON", err)
	}

	t.ID, err = tx.InsertOne(tx.SQL("tests/insert"),
		endpointID, t.Name, t.Channels, t.ChannelID, methods,
		t.Route, t.Body, data)
	if err != nil {
		return aperrors.NewWrapped("Inserting test", err)
	}

	for _, pair := range t.Pairs {
		err = pair.Insert(tx, t.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *ProxyEndpointTest) Update(tx *apsql.Tx, endpointID int64) error {
	data, err := marshaledForStorage(t.Data)
	if err != nil {
		return aperrors.NewWrapped("Marshaling test JSON", err)
	}

	methods, err := marshaledForStorage(t.Methods)
	if err != nil {
		return aperrors.NewWrapped("Marshaling test methods JSON", err)
	}

	err = tx.UpdateOne(tx.SQL("tests/update"),
		t.Name, t.Channels, t.ChannelID, methods, t.Route, t.Body,
		data, t.ID, endpointID)
	if err != nil {
		return aperrors.NewWrapped("Updating test", err)
	}

	var validPairIDs []int64
	for _, pair := range t.Pairs {
		if pair.ID == 0 {
			err = pair.Insert(tx, t.ID)
			if err != nil {
				return err
			}
		} else {
			err = pair.Update(tx, t.ID)
			if err != nil {
				return err
			}
		}
		validPairIDs = append(validPairIDs, pair.ID)
	}
	err = DeleteProxyEndpointTestPairsWithTestIDAndNotInList(tx, t.ID, validPairIDs)
	if err != nil {
		return err
	}

	return nil
}

func AllProxyEndpointTestsForEndpointID(db *apsql.DB, endpointID int64) ([]*ProxyEndpointTest, error) {
	tests := []*ProxyEndpointTest{}
	err := db.Select(&tests, db.SQL("tests/all"), endpointID)
	if err != nil {
		return nil, err
	}

	for _, test := range tests {
		test.Pairs, err = AllProxyEndpointTestPairsForTestID(db, test.ID)
		if err != nil {
			return nil, err
		}
	}

	return tests, err
}

func DeleteProxyEndpointTestsWithEndpointIDAndNotInList(tx *apsql.Tx,
	endpointID int64, validIDs []int64) error {
	args := []interface{}{endpointID}
	var validIDQuery string
	if len(validIDs) > 0 {
		validIDQuery = " AND id NOT IN (" + apsql.NQs(len(validIDs)) + ")"
		for _, id := range validIDs {
			args = append(args, id)
		}
	}
	_, err := tx.Exec(
		`DELETE FROM proxy_endpoint_tests
		WHERE endpoint_id = ?`+validIDQuery+`;`,
		args...)
	return err
}
