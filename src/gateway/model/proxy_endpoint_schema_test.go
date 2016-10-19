package model_test

import (
	"testing"

	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"
)

const JSON_SCHEMA = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "title": "Product",
    "description": "A product from Acme's catalog",
    "type": "object",
    "properties": {
        "id": {
            "description": "The unique identifier for a product",
            "type": "integer"
        }
    },
    "required": ["id"]
}`

func TestProxyEndpointSchema(t *testing.T) {
	conf := config.Database{Driver: "sqlite3", ConnectionString: ":memory:"}
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

	proxy_endpoint := &model.ProxyEndpoint{
		AccountID:     account.ID,
		UserID:        user.ID,
		APIID:         api.ID,
		EnvironmentID: environment.ID,
		Name:          "Test",
		Type:          model.ProxyEndpointTypeHTTP,
	}
	transaction("insert proxy endpoint", func(tx *apsql.Tx) error {
		return proxy_endpoint.Insert(tx)
	})

	schema := &model.ProxyEndpointSchema{
		AccountID:       account.ID,
		UserID:          user.ID,
		APIID:           api.ID,
		ProxyEndpointID: proxy_endpoint.ID,
		Name:            "Schema 1",
		RequestType:     "json",
		RequestSchema:   JSON_SCHEMA,
		ResponseType:    "json",
		ResponseSchema:  JSON_SCHEMA,
	}
	transaction("insert schema", func(tx *apsql.Tx) error {
		return schema.Insert(tx)
	})

	schemas, err := schema.All(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(schemas) != 1 {
		t.Fatal("there should be one schema")
	}

	schema.ProxyEndpointID = 0
	schemas, err = schema.All(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(schemas) != 1 {
		t.Fatal("there should be one schema")
	}
	schema.ProxyEndpointID = proxy_endpoint.ID

	schemas, err = model.FindProxyEndpointSchemasForProxy(
		db, schema.ProxyEndpointID, schema.APIID)
	if err != nil {
		t.Fatal(err)
	}
	if len(schemas) != 1 {
		t.Fatal("there should be one schema")
	}

	_, err = schema.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	transaction("update schema", func(tx *apsql.Tx) error {
		return schema.Update(tx)
	})

	transaction("delete schema", func(tx *apsql.Tx) error {
		return schema.Delete(tx)
	})
}
