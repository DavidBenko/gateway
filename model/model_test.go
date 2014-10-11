package model

import "testing"

func TestModelName(t *testing.T) {
	instance := &StructModel{}

	if instance.CollectionName() != "github.com/AnyPresence/gateway/model.StructModel" {
		t.Errorf("Expected storage key to be 'github.com/AnyPresence/gateway/model.StructModel', got %s", instance.CollectionName())
	}
}
