package model_test

import (
	"encoding/json"
	"testing"

	"gateway/config"
	"gateway/model"
	re "gateway/model/remote_endpoint"
	apsql "gateway/sql"
)

func TestMQTTSession(t *testing.T) {
	//https://groups.google.com/forum/#!topic/golang-nuts/AYZl1lNxCfA
	conf := config.Database{Driver: "sqlite3", ConnectionString: "file:mqtt_sessions.db?mode=memory&cache=shared"}
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
		Type:      model.RemoteEndpointTypePush,
	}

	platform := re.PushPlatform{
		Name:           "mqtt",
		Codename:       "mqtt",
		Type:           re.PushTypeMQTT,
		ConnectTimeout: 2,
		AckTimeout:     20,
		TimeoutRetries: 3,
	}
	push := &re.Push{
		PublishEndpoint:     true,
		SubscribeEndpoint:   true,
		UnsubscribeEndpoint: true,
		PushPlatforms:       []re.PushPlatform{platform},
	}
	data, err := json.Marshal(push)
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

	session := &model.MQTTSession{
		AccountID:        account.ID,
		APIID:            api.ID,
		RemoteEndpointID: remoteEndpoint.ID,
		Type:             platform.Codename,
		ClientID:         "abc123",
		Data:             []byte(`{"test": true}`),
	}
	transaction("insert mqtt session", func(tx *apsql.Tx) error {
		return session.Insert(tx)
	})

	all, err := session.All(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatal("there sould be 1 mqtt session")
	}

	_, err = session.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	session.Data = []byte(`{"another test": true}`)
	transaction("update mqtt sesssion", func(tx *apsql.Tx) error {
		return session.Update(tx)
	})

	count := session.Count(db)
	if count != 1 {
		t.Fatal("there sould be 1 mqtt session")
	}

	transaction("delete mqtt session", func(tx *apsql.Tx) error {
		return session.Delete(tx)
	})
}
