package model

import (
	"strconv"

	"github.com/jmoiron/sqlx/types"
	"github.com/robertkrimen/otto"
)

type Typed interface {
	GetType() string
	SetType(t string)
}

func validateJavascript(data types.JsonText, vm *otto.Otto) error {
	ds := string(data)
	if ds == "" {
		return nil
	}
	d, err := strconv.Unquote(ds)
	if err != nil {
		return err
	}
	if vm == nil {
		vm = otto.New()
	}
	_, err = vm.Compile("", d)
	if err != nil {
		return err
	}
	return nil
}
