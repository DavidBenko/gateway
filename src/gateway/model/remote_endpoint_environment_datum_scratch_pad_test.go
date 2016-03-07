package model_test

import (
	"encoding/json"
	"testing"

	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

func TestRemoteEndpointEnvironmentDatumScratchPad(t *testing.T) {
	//https://groups.google.com/forum/#!topic/golang-nuts/AYZl1lNxCfA
	conf := config.Database{Driver: "sqlite3", ConnectionString: "file:dummy.db?mode=memory&cache=shared"}
	db, err := apsql.Connect(conf)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Migrate()
	if err != nil {
		t.Fatal(err)
	}
	transaction := func(test string, operations func(tx *apsql.Tx) error) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("%v: %v", test, err)
		}
		err = operations(tx)
		if err != nil {
			t.Fatalf("%v: %v", test, err)
		}
		err = tx.Commit()
		if err != nil {
			t.Fatalf("%v: %v", test, err)
		}
	}

	account := &model.Account{
		Name: "Big Corp",
	}
	transaction("insert account", func(tx *apsql.Tx) error {
		return account.Insert(tx)
	})

	user := &model.User{
		AccountID:               account.ID,
		Name:                    "John Doe",
		Email:                   "john@bigcorp.com",
		NewPassword:             "password",
		NewPasswordConfirmation: "password",
	}
	transaction("insert user", func(tx *apsql.Tx) error {
		return user.Insert(tx)
	})

	api := &model.API{
		AccountID:          account.ID,
		UserID:             user.ID,
		Name:               "Test",
		CORSAllowOrigin:    "*",
		CORSAllowHeaders:   "*",
		CORSRequestHeaders: "*",
	}
	transaction("insert api", func(tx *apsql.Tx) error {
		return api.Insert(tx)
	})

	environment := &model.Environment{
		AccountID: account.ID,
		UserID:    user.ID,
		APIID:     api.ID,
		Name:      "default",
	}
	transaction("insert environment", func(tx *apsql.Tx) error {
		return environment.Insert(tx)
	})

	remoteEndpoint := &model.RemoteEndpoint{
		AccountID: account.ID,
		UserID:    user.ID,
		APIID:     api.ID,
		Name:      "test",
		Codename:  "test",
		Type:      model.RemoteEndpointTypeHTTP,
	}

	http := &model.HTTPRequest{
		Method: "GET",
		URL:    "http://example.com",
	}
	data, err := json.Marshal(http)
	if err != nil {
		t.Fatal(err)
	}
	remoteEndpoint.Data = data
	transaction("insert remote endpoint", func(tx *apsql.Tx) error {
		return remoteEndpoint.Insert(tx)
	})

	remoteEndpointEnvironment := &model.RemoteEndpointEnvironmentData{
		RemoteEndpointID: remoteEndpoint.ID,
		EnvironmentID:    environment.ID,
		Data:             data,
	}
	remoteEndpoint.EnvironmentData = append(remoteEndpoint.EnvironmentData, remoteEndpointEnvironment)
	transaction("update remote endpoint", func(tx *apsql.Tx) error {
		return remoteEndpoint.Update(tx)
	})

	pad := &model.RemoteEndpointEnvironmentDatumScratchPad{
		AccountID:         account.ID,
		UserID:            user.ID,
		APIID:             api.ID,
		RemoteEndpointID:  remoteEndpoint.ID,
		EnvironmentDataID: remoteEndpointEnvironment.ID,
		Name:              "test",
		Code:              "console.log('test');",
	}
	transaction("insert scratch pad", func(tx *apsql.Tx) error {
		return pad.Insert(tx)
	})

	all, err := pad.All(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatal("there sould be 1 scratch pad")
	}

	_, err = pad.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	pad.Name = "another name"
	transaction("update scratch pad", func(tx *apsql.Tx) error {
		return pad.Update(tx)
	})

	transaction("delete scratch pad", func(tx *apsql.Tx) error {
		return pad.Delete(tx)
	})
}
