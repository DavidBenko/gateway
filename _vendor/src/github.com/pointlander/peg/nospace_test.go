package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestCorrect(t *testing.T) {
	buffer := `package p
type T Peg {}
Grammar <- !.
`
	p := &Peg{Tree: New(false, false), Buffer: buffer}
	p.Init()
	err := p.Parse()
	if err != nil {
		t.Error(err)
	}
}

func TestNoSpacePackage(t *testing.T) {
	buffer := `packagenospace
type T Peg {}
Grammar <- !.
`
	p := &Peg{Tree: New(false, false), Buffer: buffer}
	p.Init()
	err := p.Parse()
	if err == nil {
		t.Error("packagenospace was parsed without error")
	}
}

func TestNoSpaceType(t *testing.T) {
	buffer := `
package p
typenospace Peg {}
Grammar <- !.
`
	p := &Peg{Tree: New(false, false), Buffer: buffer}
	p.Init()
	err := p.Parse()
	if err == nil {
		t.Error("typenospace was parsed without error")
	}
}

func TestSame(t *testing.T) {
	buffer, err := ioutil.ReadFile("peg.peg")
	if err != nil {
		t.Error(err)
	}

	p := &Peg{Tree: New(true, true), Buffer: string(buffer)}
	p.Init()
	if err := p.Parse(); err != nil {
		t.Error(err)
	}

	p.Execute()

	out := &bytes.Buffer{}
	p.Compile("peg.peg.go", out)

	bootstrap, err := ioutil.ReadFile("bootstrap.peg.go")
	if err != nil {
		t.Error(err)
	}

	if len(out.Bytes()) != len(bootstrap) {
		t.Error("code generated from peg.peg is not the same as bootstrap.peg.go")
		return
	}

	for i, v := range out.Bytes() {
		if v != bootstrap[i] {
			t.Error("code generated from peg.peg is not the same as bootstrap.peg.go")
			return
		}
	}
}
