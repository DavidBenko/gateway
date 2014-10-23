package proxy

import (
	"reflect"
	"testing"
)

func TestJoinSlices(t *testing.T) {
	query := map[string][]string{
		"foo": []string{"bar"},
		"baz": []string{"qux", "quux"},
	}
	vars := map[string][]string{
		"foo":    []string{"baz"},
		"figgle": []string{"faggle"},
	}
	expected := map[string][]string{
		"foo":    []string{"bar", "baz"},
		"baz":    []string{"qux", "quux"},
		"figgle": []string{"faggle"},
	}
	actual := joinSlices(query, vars)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected joinSlices to create %v, got %v", expected, actual)
	}
}

func TestValueInSlice(t *testing.T) {
	slice := []string{"bar"}
	if !valueInSlice("bar", slice) {
		t.Errorf("Expected bar to be in slice")
	}
	if valueInSlice("baz", slice) {
		t.Errorf("Expected baz not to be in slice")
	}
}
