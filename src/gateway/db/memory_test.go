package db

import (
	"fmt"
	"reflect"
	"testing"

	"gateway/model"
)

type testModel struct {
	id   int64
	name string `index:"true"`
}

func (t *testModel) ID() interface{} {
	return t.id
}

func (t *testModel) CollectionName() string {
	return "test_models"
}

func (t *testModel) EmptyInstance() model.Model {
	return &testModel{}
}

func (t *testModel) Valid() (bool, error) {
	return true, nil
}

func (t *testModel) Less(other model.Model) bool {
	t2 := other.(*testModel)
	return t.name < t2.name
}

func (t *testModel) MarshalToJSON(data interface{}) ([]byte, error) {
	return []byte{}, fmt.Errorf("NOPE")
}

func (t *testModel) UnmarshalFromJSON(data []byte) (model.Model, error) {
	return nil, fmt.Errorf("NOPE")
}

func (t *testModel) UnmarshalFromJSONWithID(data []byte, id interface{}) (model.Model, error) {
	return nil, fmt.Errorf("NOPE")
}

type testModel2 struct {
	name  string
	value string `index:"true" unique:"true"`
}

func (t testModel2) ID() interface{} {
	return t.name
}

func (t testModel2) CollectionName() string {
	return "test_models2"
}

func (t testModel2) EmptyInstance() model.Model {
	return testModel2{}
}

func (t testModel2) Valid() (valid bool, err error) {
	valid = t.value != "Invalid"
	if !valid {
		err = fmt.Errorf("The name cannot be 'Invalid'")
	}
	return
}

func (t testModel2) Less(instance model.Model) bool {
	return true
}

func (t testModel2) MarshalToJSON(data interface{}) ([]byte, error) {
	return []byte{}, fmt.Errorf("NOPE")
}

func (t testModel2) UnmarshalFromJSON(data []byte) (model.Model, error) {
	return nil, fmt.Errorf("NOPE")
}

func (t testModel2) UnmarshalFromJSONWithID(data []byte, id interface{}) (model.Model, error) {
	return nil, fmt.Errorf("NOPE")
}

var (
	foo = &testModel{id: 1, name: "foo"}
	bar = &testModel{id: 2, name: "bar"}

	foo2 = testModel2{name: "foo"}
	bar2 = testModel2{name: "bar"}
)

func TestSubMapInitial(t *testing.T) {
	db := NewMemoryStore()
	if len(db.storage) != 0 {
		t.Error("Expected storage to be empty initially")
	}
}

func TestSubMapPerType(t *testing.T) {
	db := NewMemoryStore()
	db.Insert(foo)
	db.Insert(bar)
	db.Insert(foo2)
	db.Insert(bar2)

	testModelSubMapsEqual := reflect.DeepEqual(db.subMap(foo), db.subMap(bar))
	testModel2SubMapsEqual := reflect.DeepEqual(db.subMap(foo2), db.subMap(bar2))
	disparateModelsEqual := reflect.DeepEqual(db.subMap(foo), db.subMap(foo2))

	if !testModelSubMapsEqual || !testModel2SubMapsEqual {
		t.Error("Expected storage to use same subMap for same types")
	}
	if disparateModelsEqual {
		t.Error("Expected storage to use different subMaps for different types")
	}
}

func TestNextID(t *testing.T) {
	db := NewMemoryStore()
	if db.NextID(&testModel{}) != int64(1) {
		t.Error("Expected empty db to have next id of 1")
	}
	db.Insert(foo)
	if db.NextID(&testModel{}) != int64(2) {
		t.Error("Expected next ID to be foo's plus one")
	}
	db.Insert(bar)
	if db.NextID(&testModel{}) != int64(3) {
		t.Error("Expected next ID to be bar's plus one")
	}
}

func TestList(t *testing.T) {
	db := NewMemoryStore()
	list, err := db.List(&testModel{})
	if err != nil {
		t.Error("Error getting list")
	}

	if len(list) != 0 {
		t.Error("Expected list to have 0 items to start")
	}

	db.Insert(foo)
	db.Insert(bar)

	list, err = db.List(&testModel{})
	if err != nil {
		t.Error("Error getting list")
	}
	if len(list) != 2 {
		t.Error("Expected list to have 2 items")
	}
	if !instanceInList(foo, list) {
		t.Error("Expected foo to be in the list")
	}
	if !instanceInList(bar, list) {
		t.Error("Expected bar to be in the list")
	}
	if list[0] != bar {
		t.Error("Expected list to be sorted alpha on name")
	}
}

func TestInsert(t *testing.T) {
	db := NewMemoryStore()
	submap := db.subMap(foo)
	count := len(submap)
	err := db.Insert(foo)
	if err != nil {
		t.Error("Expected to insert successfully")
	}
	if len(submap) != count+1 {
		t.Error("Expected Insert() to add one to the submap")
	}
	err = db.Insert(foo)
	if err == nil {
		t.Error("Expected duplicate insert to error")
	}
}

func TestInsertIndexed(t *testing.T) {
	db := NewMemoryStore()
	db.Insert(foo)
	submap := db.subMapForFieldName(foo, "name")
	if len(submap) != 1 {
		t.Error("Expected Insert() to add one to the indexed submap")
	}
}

func TestInsertUnique(t *testing.T) {
	a := testModel2{name: "a", value: "a"}
	b := testModel2{name: "b", value: "a"}
	db := NewMemoryStore()
	if err := db.Insert(a); err != nil {
		t.Error("Expected first unique insert to succeed")
	}
	if err := db.Insert(b); err == nil {
		t.Error("Expected second duplicate insert to fail")
	}
}

func TestInsertValid(t *testing.T) {
	a := testModel2{name: "a", value: "Invalid"}
	db := NewMemoryStore()
	if err := db.Insert(a); err == nil {
		t.Error("Expected invalid insert to fail")
	}
	a.value = "Valid"
	if err := db.Insert(a); err != nil {
		t.Errorf("Expected valid insert to succeed, got: %s", err)
	}
}

func TestGet(t *testing.T) {
	db := NewMemoryStore()
	_, err := db.Get(&testModel{}, 1)
	if err == nil {
		t.Error("Expected Get to return error when instance not present")
	}
	db.Insert(foo)
	instance, err := db.Get(&testModel{}, int64(1))
	if err != nil {
		t.Error("Expected Get to not return error when instance is present")
	}
	if instance != foo {
		t.Error("Expected Get to return instance requested")
	}
}

func TestFind(t *testing.T) {
	db := NewMemoryStore()
	_, err := db.Find(&testModel{}, "name", "foo")
	if err == nil {
		t.Error("Expected Find to return error when instance not present")
	}
	db.Insert(foo)
	instance, err := db.Find(&testModel{}, "name", "foo")
	if err != nil {
		t.Error("Expected Find to not return error when instance is present")
	}
	if instance != foo {
		t.Error("Expected Find to return instance requested")
	}
}

func TestUpdate(t *testing.T) {
	baz := foo

	db := NewMemoryStore()
	if err := db.Update(baz); err == nil {
		t.Error("Expected Update to return error when instance not present")
	}
	db.Insert(baz)
	baz.name = "fii"
	if err := db.Update(baz); err != nil {
		t.Error("Expected Update to not return error when instance is present")
	}
	fetched, _ := db.Get(&testModel{}, int64(1))
	if fetched.(*testModel).name != "fii" {
		t.Error("Expected Update to update the data in storage")
	}
}

func TestUpdateIndexed(t *testing.T) {
	baz := foo

	db := NewMemoryStore()
	db.Insert(baz)
	baz.name = "fii"
	db.Update(baz)
	_, err := db.Find(&testModel{}, "name", "foo")
	if err == nil {
		t.Error("Expected Update to remove old index")
	}
	instance, err := db.Find(&testModel{}, "name", "fii")
	if err != nil {
		t.Error("Expected Update to create new index")
	}
	if instance != baz {
		t.Error("Expected Find to return instance requested")
	}
}

func TestUpdateUnique(t *testing.T) {
	a := testModel2{name: "a", value: "a"}
	b := testModel2{name: "b", value: "b"}
	db := NewMemoryStore()
	db.Insert(a)
	db.Insert(b)
	if err := db.Update(b); err != nil {
		t.Error("Expected update with same unique value to succeed")
	}
	b.value = "a"
	if err := db.Update(b); err == nil {
		t.Error("Expected update to non-unique value to fail")
	}
}

func TestUpdateValid(t *testing.T) {
	a := testModel2{name: "a", value: "Valid"}
	db := NewMemoryStore()
	db.Insert(a)
	a.value = "Invalid"
	if err := db.Update(a); err == nil {
		t.Error("Expected invalid update to fail")
	}
	a.value = "Valid Again"
	if err := db.Update(a); err != nil {
		t.Errorf("Expected valid update to succeed, got %s", err)
	}
}

func TestDelete(t *testing.T) {
	db := NewMemoryStore()
	if err := db.Delete(&testModel{}, int64(1)); err == nil {
		t.Error("Expected Delete to return error when instance not present")
	}
	db.Insert(foo)
	if err := db.Delete(&testModel{}, int64(1)); err != nil {
		t.Error("Expected Delete to not return error when instance is present")
	}
	if _, err := db.Get(&testModel{}, int64(1)); err == nil {
		t.Error("Expected Delete to remove the instance from storage")
	}
}

func TestDeleteIndexed(t *testing.T) {
	db := NewMemoryStore()
	db.Insert(foo)
	db.Delete(&testModel{}, int64(1))
	_, err := db.Find(&testModel{}, "name", "foo")
	if err == nil {
		t.Error("Expected Delete to remove old index")
	}
}

func TestDeleteUnique(t *testing.T) {
	a := testModel2{name: "a", value: "a"}
	b := testModel2{name: "b", value: "b"}
	db := NewMemoryStore()
	db.Insert(a)
	db.Insert(b)
	b.value = "a"
	if err := db.Update(b); err == nil {
		t.Error("Expected update to non-unique value to fail")
	}
	db.Delete(testModel2{}, a.ID())
	if err := db.Update(b); err != nil {
		t.Error("Expected update to now-unique value to succeed")
	}
}

func instanceInList(a model.Model, list []model.Model) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
