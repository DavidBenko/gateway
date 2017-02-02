package repl

import (
	"encoding/json"
	"gateway/logreport"

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
		logreport.Println(err)
		ef := &Frame{Data: "unknown error occurred", Type: ERROR}
		return ef.JSON()
	}
	return out
}

func NewRepl(vm *otto.Otto) *Repl {
	output := make(chan []byte)
	input := make(chan []byte)
	r := &Repl{vm, input, output, make(chan bool)}
	return r
}

func (r *Repl) Run() {
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
	return
}

func (r *Repl) Stop() {
	r.stop <- true
}
