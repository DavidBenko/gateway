package http

import (
	"fmt"
	"testing"
)

func TestHTTPErrorDefault500(t *testing.T) {
	err := httpError{}
	if err.Code() != 500 {
		t.Error("Expected default error to be 500")
	}
}

func TestHTTPErrorCustom(t *testing.T) {
	err := httpError{code: 400}
	if err.Code() != 400 {
		t.Error("Expected custom error code to be used")
	}
}

func TestNewServerError(t *testing.T) {
	err := NewServerError(fmt.Errorf("%s", "error"))
	if err.Code() != 500 {
		t.Error("Expected default error to be 500")
	}
	if err.Error().Error() != "error" {
		t.Error("Expected custom error to be used")
	}
}

func TestNewError(t *testing.T) {
	err := NewError(fmt.Errorf("%s", "error"), 400)
	if err.Code() != 400 {
		t.Error("Expected custom error code to be used")
	}
	if err.Error().Error() != "error" {
		t.Error("Expected custom error to be used")
	}
}
