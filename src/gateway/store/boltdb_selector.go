package store

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

type Selector struct {
	context
	json  interface{}
	param []interface{}
}

type Value struct {
	b      bool
	errors []error
}

func (s *Selector) process(node *node32) (v Value) {
	for node != nil {
		switch node.pegRule {
		case rulee:
			return s.process(node.up)
		case rulee1:
			v = s.processRulee1(node.up)
		}
		node = node.next
	}
	return
}

func (s *Selector) processRulee1(node *node32) (v Value) {
	for node != nil {
		if node.pegRule == rulee2 {
			if !v.b {
				x := s.processRulee2(node.up)
				v.b = v.b || x.b
				v.errors = append(v.errors, x.errors...)
			}
		}
		node = node.next
	}
	return
}

func (s *Selector) processRulee2(node *node32) (v Value) {
	v.b = true
	for node != nil {
		if node.pegRule == rulee3 {
			if v.b {
				x := s.processRulee3(node.up)
				v.b = v.b && x.b
				v.errors = append(v.errors, x.errors...)
			}
		}
		node = node.next
	}
	return
}

func (s *Selector) processRulee3(node *node32) (v Value) {
	if node.pegRule == ruleexpression {
		return s.processExpression(node.up)
	}
	return s.process(node.next.up)
}

func (s *Selector) processExpression(node *node32) (v Value) {
	if node.pegRule == ruleboolean {
		v.b = s.Node(node) == "true"
		return
	}

	path, _json, valid := node.up, s.json, false
	for path != nil {
		if path.pegRule == ruleword {
			_json, valid = _json.(map[string]interface{})[s.Node(path)]
			if !valid {
				return
			}
		}
		path = path.next
	}
	node = node.next
	op := s.Node(node)
	node = node.next.up
	_a := fmt.Sprintf("%v", _json)
	switch node.pegRule {
	case ruleplaceholder:
		placeholder, err := strconv.Atoi(string(s.buffer[node.begin+1 : node.end]))
		if err != nil {
			v.errors = append(v.errors, err)
			return
		}

		if placeholder > len(s.param) {
			v.errors = append(v.errors, errors.New("placholder to large"))
			return
		}
		switch _b := s.param[placeholder-1].(type) {
		case string:
			v.b = compareString(op, _a, _b)
		case float64:
			a, b := &big.Rat{}, &big.Rat{}
			a.SetString(_a)
			b.SetFloat64(_b)
			v.b = compareRat(op, a, b)
		case int:
			a, b := &big.Rat{}, &big.Rat{}
			a.SetString(_a)
			b.SetInt64(int64(_b))
			v.b = compareRat(op, a, b)
		case bool:
			v.b = compareBool(op, _a == "true", _b)
		default:
			v.b = compareNull(op, _json, _b)
		}
	case rulestring:
		b := string(s.buffer[node.begin+1 : node.end-1])
		v.b = compareString(op, _a, b)
	case rulenumber:
		_b := s.Node(node)
		b := &big.Rat{}
		b.SetString(_b)
		a := &big.Rat{}
		a.SetString(_a)
		v.b = compareRat(op, a, b)
	case ruleboolean:
		b := s.Node(node)
		v.b = compareBool(op, _a == "true", b == "true")
	case rulenull:
		v.b = compareNull(op, _json, nil)
	}
	return
}

func compareString(op, a, b string) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	}
	return false
}

func compareRat(op string, a, b *big.Rat) bool {
	switch op {
	case "=":
		return a.Cmp(b) == 0
	case "!=":
		return a.Cmp(b) != 0
	case ">":
		return a.Cmp(b) > 0
	case "<":
		return a.Cmp(b) < 0
	case ">=":
		return a.Cmp(b) >= 0
	case "<=":
		return a.Cmp(b) <= 0
	}
	return false
}

func compareBool(op string, a, b bool) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	}
	return false
}

func compareNull(op string, a, b interface{}) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	}
	return false
}
