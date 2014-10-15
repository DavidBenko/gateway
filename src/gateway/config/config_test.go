package config

import (
	"reflect"
	"testing"
)

type testConfig struct {
	Foo string `flag:"foo" default:""`
	Bar int64  `flag:"bar" default:"42"`
	Baz baz
}

type baz struct {
	Baf string `flag:"baf" default:"doby"`
}

func TestSetDefaults(t *testing.T) {
	config := testConfig{}
	setDefaults(reflect.ValueOf(&config).Elem())
	if config.Foo != "" {
		t.Errorf("Expected default foo value to be set.")
	}
	if config.Bar != 42 {
		t.Errorf("Expected default bar value to be set.")
	}
	if config.Baz.Baf != "doby" {
		t.Errorf("Expected default baz value to be set.")
	}
}
