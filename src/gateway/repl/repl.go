package repl

import (
	"errors"
	"io"

	"github.com/robertkrimen/otto"
)

type Repl struct {
	vm        *otto.Otto
	Rwc       io.ReadWriteCloser
	accountID int64
}

func NewRepl(vm *otto.Otto, rwc io.ReadWriteCloser, accountID int64) (*Repl, error) {
	if accountID == 0 {
		return nil, errors.New("invalid accountID 0")
	}
	r := &Repl{vm, rwc, accountID}
	return r, nil
}

func (r *Repl) Start() error {
	return nil
}
