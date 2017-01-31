package repl

import (
	"encoding/json"

	"github.com/robertkrimen/otto"
)

const (
	OUTPUT   = "output"
	ERROR    = "error"
	HEARBEAT = "heartbeat"
)

type Repl struct {
	vm     *otto.Otto
	Input  chan []byte
	Output chan []byte
	stop   chan bool
}

type Frame struct {
	Data string `json:"data,omitempty"`
	Type string `json:"type"`
}

func (f *Frame) JSON() []byte {
	out, err := json.Marshal(f)
	if err != nil {
		// create JSON error
	}
	return out
}

func NewRepl(vm *otto.Otto, input chan []byte) (*Repl, error) {
	output := make(chan []byte)
	r := &Repl{vm, input, output, make(chan bool, 1)}
	return r, nil
}

func (r *Repl) Run() error {
l:
	for {
		select {
		case in := <-r.Input:
			val, err := r.vm.Run(in)
			if err != nil {
				frame := &Frame{err.Error(), ERROR}
				r.Output <- frame.JSON()
				break
			}
			frame := &Frame{val.String(), OUTPUT}
			r.Output <- frame.JSON()
		case <-r.stop:
			break l
		}
	}
	return nil
}

func (r *Repl) Stop() {
	r.stop <- true
}
