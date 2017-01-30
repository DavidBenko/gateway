package repl

import "github.com/robertkrimen/otto"

type Repl struct {
	vm     *otto.Otto
	input  chan []byte
	Output chan []byte
	stop   chan bool
}

func NewRepl(vm *otto.Otto, input chan []byte) (*Repl, error) {
	output := make(chan []byte)
	r := &Repl{vm, input, output, make(chan bool, 1)}
	return r, nil
}

func (r *Repl) Start() error {
	r.Output <- []byte("foo bar baz!")
	// TODO enter loop for reading/writing to channels
	return nil
}

func (r *Repl) Stop() {
	r.stop <- true
}
