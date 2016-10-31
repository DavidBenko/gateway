package store_test

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"gateway/config"
	"gateway/store"

	"github.com/ory-am/dockertest"
)

var testStore store.Store

func testPostgres(m *testing.M) int {
	fmt.Println("testPostgres")
	var postgresStore store.Store
	conf := config.Store{
		Type: "postgres",
	}
	c, err := dockertest.ConnectToPostgreSQL(60, time.Second, func(url string) bool {
		conf.ConnectionString = url
		postgresStore, _ = store.Configure(conf)
		return postgresStore.(*store.PostgresStore).Ping() == nil
	})
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}
	defer func() {
		postgresStore.Shutdown()
		c.KillRemove()
	}()
	testStore = postgresStore
	return m.Run()
}

func testBolt(m *testing.M) int {
	fmt.Println("testBolt")
	var boltStore store.Store
	file := make([]byte, 8)
	binary.BigEndian.PutUint64(file, uint64(rand.Int63()))
	name := os.TempDir() + string(os.PathSeparator) + hex.EncodeToString(file) + ".db"
	conf := config.Store{
		Type:             "boltdb",
		ConnectionString: name,
	}
	boltStore, err := store.Configure(conf)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		boltStore.Shutdown()
		err := os.Remove(name)
		if err != nil {
			log.Fatal(err)
		}
	}()
	testStore = boltStore
	return m.Run()
}

func TestMain(m *testing.M) {
	flag.Parse()

	status := testPostgres(m)
	if status != 0 {
		os.Exit(status)
	}

	status = testBolt(m)
	os.Exit(status)
}

var testJson = []string{
	`{
		"name": {
			"first": "John",
			"last": "Doe"
		},
		"age": 25,
		"salary": 51
	}`,
	`{
		"name": {
			"first": "Jane",
			"last": "Doe"
		},
		"age": 21,
		"salary": 43
	}`,
	`{
		"name": {
			"first": "Joe",
			"last": "Schmo"
		},
		"age": 18,
		"salary": 37
	}`,
}

func setup(t *testing.T) store.Store {
	err := testStore.Migrate()
	if err != nil {
		t.Fatal(err)
	}
	return testStore
}

func teardown(t *testing.T, s store.Store) {
	err := s.Clear()
	if err != nil {
		t.Fatal(err)
	}
}

func parse(t *testing.T) []interface{} {
	_json := []interface{}{}
	for _, test := range testJson {
		var parsed interface{}
		err := json.Unmarshal([]byte(test), &parsed)
		if err != nil {
			t.Fatal(err)
		}
		_json = append(_json, parsed)
	}
	return _json
}

func TestConfigure(t *testing.T) {
	s := setup(t)
	teardown(t, s)
}

func TestCreateCollection(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}
}

func TestListCollection(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}
	collections := []*store.Collection{}
	err = s.ListCollection(&store.Collection{AccountID: 0}, &collections)
	if err != nil {
		t.Fatal(err)
	}
	if len(collections) != 1 {
		t.Fatal("there should be 1 collection")
	}
}

func TestShowCollection(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}

	err = s.ShowCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.Name == "" {
		t.Fatal("collection name should be set")
	}
}

func TestUpdateCollection(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}

	collection.Name = "anewname"
	err = s.UpdateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCollection(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}

	err = s.DeleteCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateObject(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}

	object := &store.Object{
		AccountID:    0,
		CollectionID: collection.ID,
		Data:         []byte(`{"test": "object"}`),
	}
	err = s.CreateObject(object)
	if err != nil {
		log.Fatal(err)
	}
	if object.ID == 0 {
		t.Fatal("failed to create object")
	}
}

func TestListObject(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}

	object := &store.Object{
		AccountID:    0,
		CollectionID: collection.ID,
		Data:         []byte(`{"test": "object"}`),
	}
	err = s.CreateObject(object)
	if err != nil {
		log.Fatal(err)
	}
	if object.ID == 0 {
		t.Fatal("failed to create object")
	}

	var objects []*store.Object
	err = s.ListObject(object, &objects)
	if err != nil {
		log.Fatal(err)
	}
	if len(objects) != 1 {
		t.Fatal("there should be 1 objects")
	}
}

func TestShowObject(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}

	object := &store.Object{
		AccountID:    0,
		CollectionID: collection.ID,
		Data:         []byte(`{"test": "object"}`),
	}
	err = s.CreateObject(object)
	if err != nil {
		log.Fatal(err)
	}
	if object.ID == 0 {
		t.Fatal("failed to create object")
	}

	err = s.ShowObject(object)
	if err != nil {
		log.Fatal(err)
	}
	if object.Data == nil {
		t.Fatal("object should have data")
	}
}

func TestUpdateObject(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}

	object := &store.Object{
		AccountID:    0,
		CollectionID: collection.ID,
		Data:         []byte(`{"test": "object"}`),
	}
	err = s.CreateObject(object)
	if err != nil {
		log.Fatal(err)
	}
	if object.ID == 0 {
		t.Fatal("failed to create object")
	}

	object.Data = []byte(`{"test": "another object"}`)
	err = s.UpdateObject(object)
	if err != nil {
		log.Fatal(err)
	}
}

func TestDeleteObject(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	collection := &store.Collection{AccountID: 0, Name: "acollection"}
	err := s.CreateCollection(collection)
	if err != nil {
		t.Fatal(err)
	}
	if collection.ID == 0 {
		t.Fatal("failed to create collection")
	}

	object := &store.Object{
		AccountID:    0,
		CollectionID: collection.ID,
		Data:         []byte(`{"test": "object"}`),
	}
	err = s.CreateObject(object)
	if err != nil {
		log.Fatal(err)
	}
	if object.ID == 0 {
		t.Fatal("failed to create object")
	}

	err = s.DeleteObject(object)
	if err != nil {
		log.Fatal(err)
	}
}

func TestInsert(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	objects := parse(t)
	objects, err := s.Insert(0, "people", objects[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, object := range objects {
		if object.(map[string]interface{})["$id"] == nil {
			t.Fatal("object $id should be set")
		}
	}
}

func TestInsertArray(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	objects := parse(t)
	objs, err := s.Insert(0, "people", objects)
	if err != nil {
		t.Fatal(err)
	}
	for _, object := range objs {
		if object.(map[string]interface{})["$id"] == nil {
			t.Fatal("object $id should be set")
		}
	}
}

func TestSelectByID(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	objects := parse(t)
	objects, err := s.Insert(0, "people", objects[0])
	if err != nil {
		t.Fatal(err)
	}
	object := objects[0]
	if object.(map[string]interface{})["$id"] == nil {
		t.Fatal("object $id should be set")
	}
	id := object.(map[string]interface{})["$id"].(uint64)
	object, err = s.SelectByID(0, "people", id)
	if err != nil {
		t.Fatal(err)
	}
	if object.(map[string]interface{})["$id"] == nil {
		t.Fatal("object $id should be set")
	}
}

func TestUpdateByID(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	objects := parse(t)
	objects, err := s.Insert(0, "people", objects[0])
	if err != nil {
		t.Fatal(err)
	}
	object := objects[0]
	if object.(map[string]interface{})["$id"] == nil {
		t.Fatal("object $id should be set")
	}
	_json := object.(map[string]interface{})
	id := _json["$id"].(uint64)
	_json["age"] = 26
	object, err = s.UpdateByID(0, "people", id, object)
	if err != nil {
		t.Fatal(err)
	}
	if object.(map[string]interface{})["$id"] == nil {
		t.Fatal("object $id should be set")
	}
}

func TestDeleteByID(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	objects := parse(t)
	objects, err := s.Insert(0, "people", objects[0])
	if err != nil {
		t.Fatal(err)
	}
	object := objects[0]
	if object.(map[string]interface{})["$id"] == nil {
		t.Fatal("object $id should be set")
	}
	id := object.(map[string]interface{})["$id"].(uint64)
	object, err = s.DeleteByID(0, "people", id)
	if err != nil {
		t.Fatal(err)
	}
	if object.(map[string]interface{})["$id"] == nil {
		t.Fatal("object $id should be set")
	}
}

func TestDeleteBulk(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	objects := parse(t)
	objs, err := s.Insert(0, "people", objects)
	if err != nil {
		t.Fatal(err)
	}
	for _, object := range objs {
		if object.(map[string]interface{})["$id"] == nil {
			t.Fatal("object $id should be set")
		}
	}

	objs, err = s.Delete(0, "people", "age >= $1", 18)
	if err != nil {
		t.Fatal(err)
	}
	if len(objs) != 3 {
		t.Fatal("there should be 3 objects")
	}
	for _, object := range objs {
		if object.(map[string]interface{})["$id"] == nil {
			t.Fatal("object $id should be set")
		}
	}
}

func TestSelect(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	objects := parse(t)
	for _, obj := range objects {
		objects, err := s.Insert(0, "people", obj)
		if err != nil {
			t.Fatal(err)
		}
		object := objects[0]
		if object.(map[string]interface{})["$id"] == nil {
			t.Fatal("object $id should be set")
		}
	}
	objects, err := s.Select(0, "people", "age >= 21 order numeric(age) asc")
	t.Log(objects)
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 2 {
		t.Fatal("there should be 2 objects")
	}
	for _, object := range objects {
		if object.(map[string]interface{})["$id"] == nil {
			t.Fatal("object $id should be set")
		}
	}
	last := int64(0)
	for _, object := range objects {
		age := object.(map[string]interface{})["age"].(float64)
		if int64(age) < last {
			t.Fatal("objects should be sorted")
		}
		last = int64(age)
	}

	objects, err = s.Select(0, "people", "true")
	t.Log(objects)
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 3 {
		t.Fatal("there should be 3 objects")
	}
}

func TestSelectUnderscore(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	test :=
		`{
			"_name_": {
				"_first_": "John",
				"_last_": "Doe"
			},
			"_age_": 25
		}`
	var parsed interface{}
	err := json.Unmarshal([]byte(test), &parsed)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Insert(0, "_people_", parsed)
	if err != nil {
		t.Fatal(err)
	}

	objects, err := s.Select(0, "_people_", "_age_ >= 21 order numeric(_age_) asc")
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 1 {
		t.Fatal("there should be 1 object")
	}
}

func TestAggregation(t *testing.T) {
	s := setup(t)
	defer teardown(t, s)

	objects := parse(t)
	for _, obj := range objects {
		objects, err := s.Insert(0, "people", obj)
		if err != nil {
			t.Fatal(err)
		}
		object := objects[0]
		if object.(map[string]interface{})["$id"] == nil {
			t.Fatal("object $id should be set")
		}
	}
	objects, err := s.Select(0, "people", `true | count(name.first) as count,
		count(*) as countall, sum(age) as sum, avg(age) as avg,
		stddev(age) as stddev, min(age) as min, max(age) as max,
		corr(age, salary) as corr, cov(age, salary) as cov, var(age) as var`)
	t.Log(objects)
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 1 {
		t.Fatal("there should be 1 objects")
	}
	valid := map[string]string{
		"count":    "3.00",
		"countall": "3.00",
		"sum":      "64.00",
		"avg":      "21.33",
		"stddev":   "2.87",
		"min":      "18.00",
		"max":      "25.00",
		"corr":     "1.00",
		"cov":      "16.44",
		"var":      "8.22",
	}
	results := objects[0].(map[string]interface{})
	for name, value := range valid {
		if s := fmt.Sprintf("%.2f", results[name].(float64)); s != value {
			t.Fatal(fmt.Sprintf("%v should be equal to %v not %v", name, value, s))
		}
	}
}
