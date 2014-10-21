package proxy

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
	actual := desliceValues(slice)
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
	actual := resliceValues(slice)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected desliceValues to convert singe value slice to value")
	}
}

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
