package repl

import (
	"fmt"

	"github.com/robertkrimen/otto"
)

type Repl struct {
	vm     *otto.Otto
	Input  chan []byte
	Output chan []byte
	stop   chan bool
}

func NewRepl(vm *otto.Otto, input chan []byte) (*Repl, error) {
	output := make(chan []byte)
	r := &Repl{vm, input, output, make(chan bool, 1)}
	return r, nil
}

func (r *Repl) Run() {
l:
	for {
		select {
		case in := <-r.Input:
			val, err := r.vm.Run(in)
			if err != nil {
				r.Output <- []byte(fmt.Sprintf("error: %s", err.Error()))
				break
			}
			r.Output <- []byte(val.String())
		case <-r.stop:
			break l
		}
	}
}

func (r *Repl) Stop() {
	r.stop <- true
}
