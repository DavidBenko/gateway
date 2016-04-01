package model_test

import (
	"encoding/json"
	"testing"

	"gateway/config"
	"gateway/model"
	re "gateway/model/remote_endpoint"
	apsql "gateway/sql"
)

func TestPush(t *testing.T) {
	//https://groups.google.com/forum/#!topic/golang-nuts/AYZl1lNxCfA
	conf := config.Database{Driver: "sqlite3", ConnectionString: "file:push_channels.db?mode=memory&cache=shared"}
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
		Name:     "Test",
		Codename: "test",
		Type:     re.PushTypeGCM,
		APIKey:   "AIzaSyCPc5PN7PkKT7BGj-b60XAmEpp5f9N1oNY",
	}
	data, err := json.Marshal(&re.Push{[]re.PushPlatform{platform}})
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

	channel := &model.PushChannel{
		AccountID:        account.ID,
		UserID:           user.ID,
		APIID:            api.ID,
		RemoteEndpointID: remoteEndpoint.ID,
		Name:             "test",
	}
	transaction("insert push channel", func(tx *apsql.Tx) error {
		return channel.Insert(tx)
	})

	all, err := channel.All(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatal("there sould be 1 push channel")
	}

	_, err = channel.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	channel.Name = "another name"
	transaction("update push channel", func(tx *apsql.Tx) error {
		return channel.Update(tx)
	})

	device := &model.PushDevice{
		AccountID:        account.ID,
		UserID:           user.ID,
		APIID:            api.ID,
		RemoteEndpointID: remoteEndpoint.ID,
		Name:             "test",
		PushChannelID:    channel.ID,
		Type:             re.PushTypeGCM,
		Token:            "abc123",
	}
	transaction("insert push device", func(tx *apsql.Tx) error {
		return device.Insert(tx)
	})

	allDevices, err := device.All(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(allDevices) != 1 {
		t.Fatal("there sould be 1 push device")
	}

	_, err = device.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	device.Name = "another name"
	transaction("update push device", func(tx *apsql.Tx) error {
		return device.Update(tx)
	})

	message := &model.PushMessage{
		AccountID:        account.ID,
		UserID:           user.ID,
		APIID:            api.ID,
		RemoteEndpointID: remoteEndpoint.ID,
		PushChannelID:    channel.ID,
		PushDeviceID:     device.ID,
		Stamp:            123,
	}
	transaction("insert push message", func(tx *apsql.Tx) error {
		return message.Insert(tx)
	})

	allMessages, err := device.All(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(allMessages) != 1 {
		t.Fatal("there sould be 1 push message")
	}

	_, err = message.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	message.Stamp = 456
	transaction("update push message", func(tx *apsql.Tx) error {
		return message.Update(tx)
	})

	transaction("delete push message", func(tx *apsql.Tx) error {
		return message.Delete(tx)
	})

	transaction("delete push device", func(tx *apsql.Tx) error {
		return device.Delete(tx)
	})

	transaction("delete push channel", func(tx *apsql.Tx) error {
		return channel.Delete(tx)
	})
}
