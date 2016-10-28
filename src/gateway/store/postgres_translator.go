package store

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Translator struct {
	context
	param []interface{}
}

type Query struct {
	s, aggregate string
	errors       []error
}

func (t *Translator) Process(node *node32) (q Query) {
	for node != nil {
		switch node.pegRule {
		case rulee:
			return t.Process(node.up)
		case rulee1:
			x := t.ProcessRulee1(node.up)
			q.s += "( " + x.s + " )"
			q.errors = append(q.errors, x.errors...)
		case ruleorder:
			x := t.ProcessOrder(node.up)
			q.s += " " + x.s
		case rulelimit:
			x := t.ProcessLimit(node.up)
			q.s += " " + x.s
		case ruleoffset:
			x := t.ProcessOffset(node.up)
			q.s += " " + x.s
		case ruleaggregate:
			x := t.ProcessAggregate(node.up)
			q.aggregate = x.aggregate
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessOrder(node *node32) (q Query) {
	for node != nil {
		switch node.pegRule {
		case rulepath:
			path := t.ProcessPath(node.up)
			q.s = "ORDER BY " + path.s
		case rulecast:
			cast := t.ProcessCast(node.up)
			q.s = "ORDER BY " + cast.s
		case ruleasc:
			q.s += " ASC"
		case ruledesc:
			q.s += " DESC"
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessCast(node *node32) (q Query) {
	for node != nil {
		if node.pegRule == rulepath {
			path := t.ProcessPath(node.up)
			q.s = "CAST( " + path.s + " as numeric )"
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessPath(node *node32) (q Query) {
	segments := []string{}
	for node != nil {
		if node.pegRule == ruleword {
			segments = append(segments, t.Node(node))
		}
		node = node.next
	}
	last := len(segments) - 1
	q.s = "data"
	for _, segment := range segments[:last] {
		q.s += "->'" + segment + "'"
	}
	q.s += "->>'" + segments[last] + "'"
	return
}

func (t *Translator) ProcessLimit(node *node32) (q Query) {
	for node != nil {
		if node.pegRule == rulevalue1 {
			q.s += "LIMIT " + t.Node(node)
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessOffset(node *node32) (q Query) {
	for node != nil {
		if node.pegRule == rulevalue1 {
			q.s += "OFFSET " + t.Node(node)
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessAggregate(node *node32) (q Query) {
	comma := ""
	for node != nil {
		if node.pegRule == ruleaggregate_clause {
			x := t.ProcessAggregateClause(node.up)
			q.aggregate += comma + " " + x.aggregate
			comma = ","
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessAggregateClause(node *node32) (q Query) {
	function := ""
	for node != nil {
		switch node.pegRule {
		case rulefunction:
			function = t.Node(node)
			if function == "stddev" {
				function = "stddev_pop"
			}
			q.aggregate += function + "( "
		case ruleselector:
			selector := t.ProcessSelector(node.up).aggregate
			if function == "count" {
				q.aggregate += selector + " )"
			} else {
				q.aggregate += "CAST( " + selector + "as float ) )"
			}
		case ruleword:
			q.aggregate += " as " + t.Node(node)
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessSelector(node *node32) (q Query) {
	for node != nil {
		switch node.pegRule {
		case rulepath:
			x := t.ProcessPath(node.up)
			q.aggregate = x.s
		case rulewildcard:
			q.aggregate = t.Node(node)
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessRulee1(node *node32) (q Query) {
	or := ""
	for node != nil {
		if node.pegRule == rulee2 {
			x := t.ProcessRulee2(node.up)
			q.s += or + x.s
			q.errors = append(q.errors, x.errors...)
			or = " OR "
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessRulee2(node *node32) (q Query) {
	and := ""
	for node != nil {
		if node.pegRule == rulee3 {
			x := t.ProcessRulee3(node.up)
			q.s += and + x.s
			q.errors = append(q.errors, x.errors...)
			and = " AND "
		}
		node = node.next
	}
	return
}

func (t *Translator) ProcessRulee3(node *node32) (q Query) {
	if node.pegRule == ruleexpression {
		return t.ProcessExpression(node.up)
	}
	x := t.Process(node.next.up)
	q.s = "(" + x.s + ")"
	q.errors = x.errors
	return
}

func (t *Translator) ProcessExpression(node *node32) (q Query) {
	if node.pegRule == ruleboolean {
		q.s = t.Node(node)
		return
	}

	path, segments := node.up, []string{}
	for path != nil {
		if path.pegRule == ruleword {
			segments = append(segments, t.Node(path))
		}
		path = path.next
	}
	q.s = "data"
	last := len(segments) - 1
	for _, segment := range segments[:last] {
		q.s += "->'" + segment + "'"
	}
	q.s += "->>'" + segments[last] + "'"

	node = node.next
	op := t.Node(node)
	node = node.next.up
	switch node.pegRule {
	case ruleplaceholder:
		placeholder, err := strconv.Atoi(string(t.buffer[node.begin+1 : node.end]))
		if err != nil {
			q.errors = append(q.errors, err)
			return
		}

		if placeholder > len(t.param) {
			q.errors = append(q.errors, errors.New("placholder to large"))
			return
		}
		switch t.param[placeholder-1].(type) {
		case string:
			q.s = fmt.Sprintf("%v %v $%v", q.s, op, placeholder)
		case float64:
			q.s = fmt.Sprintf("CAST(%v as FLOAT) %v $%v", q.s, op, placeholder)
		case int:
			q.s = fmt.Sprintf("CAST(%v as INTEGER) %v $%v", q.s, op, placeholder)
		case bool:
			q.s = fmt.Sprintf("CAST(%v as BOOLEAN) %v $%v", q.s, op, placeholder)
		default:
			switch op {
			case "=":
				q.s = fmt.Sprintf("%v IS NULL", q.s)
			case "!=":
				q.s = fmt.Sprintf("%v IS NOT NULL", q.s)
			}
		}
	case rulestring:
		param := string(t.buffer[node.begin+1 : node.end-1])
		q.s = fmt.Sprintf("%v %v '%v'", q.s, op, param)
	case rulenumber:
		param := t.Node(node)
		if strings.Contains(param, ".") {
			q.s = fmt.Sprintf("CAST(%v as FLOAT) %v %v", q.s, op, param)
		} else {
			q.s = fmt.Sprintf("CAST(%v as INTEGER) %v %v", q.s, op, param)
		}
	case ruleboolean:
		param := t.Node(node)
		q.s = fmt.Sprintf("CAST(%v as BOOLEAN) %v %v", q.s, op, param)
	case rulenull:
		switch op {
		case "=":
			q.s = fmt.Sprintf("%v IS NULL", q.s)
		case "!=":
			q.s = fmt.Sprintf("%v IS NOT NULL", q.s)
		}
	}
	return
}
