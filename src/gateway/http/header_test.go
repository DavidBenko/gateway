package http

import (
	"reflect"
	"testing"
)

func TestDesliceValues(t *testing.T) {
	slice := map[string][]string{
		"foo": []string{"bar"},
		"baz": []string{"qux", "quux"},
	}
	expected := map[string]interface{}{
		"foo": "bar",
		"baz": []string{"qux", "quux"},
	}
	actual := DesliceValues(slice)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected desliceValues to convert singe value slice to value")
	}
}

func TestResliceValues(t *testing.T) {
	slice := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	expected := map[string][]string{
		"foo": []string{"bar"},
		"baz": []string{"qux"},
	}
	actual := ResliceValues(slice)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected desliceValues to convert singe value slice to value")
	}
}
