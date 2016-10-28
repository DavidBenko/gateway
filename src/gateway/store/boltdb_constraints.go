package store

import (
	"errors"
	"strconv"
)

type Constraints struct {
	context
	param  []interface{}
	errors []error
	order  struct {
		path    []string
		dir     string
		numeric bool
	}
	hasLimit  bool
	limit     int
	hasOffset bool
	offset    int
}

func (c *Constraints) process(node *node32) {
	for node != nil {
		switch node.pegRule {
		case rulee:
			c.process(node.up)
		case ruleorder:
			c.processOrder(node.up)
		case rulelimit:
			c.processLimit(node.up)
		case ruleoffset:
			c.processOffset(node.up)
		}
		node = node.next
	}
	return
}

func (c *Constraints) processOrder(node *node32) {
	for node != nil {
		switch node.pegRule {
		case rulepath:
			c.processPath(node.up)
		case rulecast:
			c.processCast(node.up)
		case ruleasc:
			c.order.dir = c.Node(node)
		case ruledesc:
			c.order.dir = c.Node(node)
		}
		node = node.next
	}
}

func (c *Constraints) processCast(node *node32) {
	c.order.numeric = true
	for node != nil {
		if node.pegRule == rulepath {
			c.processPath(node.up)
		}
		node = node.next
	}
}

func (c *Constraints) processPath(node *node32) {
	for node != nil {
		if node.pegRule == ruleword {
			c.order.path = append(c.order.path, c.Node(node))
		}
		node = node.next
	}
}

func (c *Constraints) processLimit(node *node32) {
	c.hasLimit = true
	for node != nil {
		if node.pegRule == rulevalue1 {
			var err error
			c.limit, err = c.processValue1(node.up)
			if err != nil {
				c.errors = append(c.errors, err)
			}
		}
		node = node.next
	}
}

func (c *Constraints) processOffset(node *node32) {
	c.hasOffset = true
	for node != nil {
		if node.pegRule == rulevalue1 {
			var err error
			c.offset, err = c.processValue1(node.up)
			if err != nil {
				c.errors = append(c.errors, err)
			}
		}
		node = node.next
	}
}

func (c *Constraints) processValue1(node *node32) (int, error) {
	for node != nil {
		switch node.pegRule {
		case ruleplaceholder:
			placeholder, err := strconv.Atoi(string(c.buffer[node.begin+1 : node.end]))
			if err != nil {
				return -1, err
			}
			if placeholder > len(c.param) {
				return -1, errors.New("placholder too large")
			}
			if holder, valid := c.param[placeholder-1].(int); valid {
				return holder, nil
			}
			return -1, errors.New("value must be type int")
		case rulewhole:
			whole, err := strconv.Atoi(c.Node(node))
			if err != nil {
				return -1, err
			}
			return whole, nil
		}
		node = node.next
	}
	return -1, errors.New("no value")
}
