package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"gateway/config"

	"github.com/boltdb/bolt"
)

const (
	StoreTypeBolt     = "boltdb"
	StoreTypePostgres = "postgres"
)

type Store struct {
	conf   config.Store
	boltdb *bolt.DB
}

func Configure(conf config.Store) (*Store, error) {
	s := Store{conf: conf}
	if conf.Type == StoreTypeBolt {
		var err error
		s.boltdb, err = bolt.Open(conf.ConnectionString, 0600, nil)
		if err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func (s *Store) Select(collection string, query string, params ...interface{}) (error, []interface{}) {
	tx, err := s.boltdb.Begin(false)
	if err != nil {
		return err, nil
	}
	defer tx.Rollback()
	bucket := tx.Bucket([]byte(collection))
	if bucket == nil {
		return errors.New("collection deosn't exist"), nil
	}

	jql := &JQL{Buffer: query}
	jql.Init()
	err = jql.Parse()
	if err != nil {
		return err, nil
	}

	ast, buffer := jql.tokenTree.AST(), []rune(jql.Buffer)
	constraints := getConstraints(ast, &Context{buffer, nil, params})
	var results []interface{}
	if len(constraints.order.path) > 0 {
		cursor := bucket.Cursor()
		key, value := cursor.Next()
		for key != nil {
			_value := make([]byte, len(value))
			copy(_value, value)
			decoder := json.NewDecoder(bytes.NewReader(_value))
			decoder.UseNumber()
			var _json interface{}
			err = decoder.Decode(&_json)
			if err != nil {
				return err, nil
			}
			if process(ast, &Context{buffer, _json, params}).b {
				_key := make([]byte, len(key))
				copy(_key, key)
				_json.(map[string]interface{})["$id"] = string(_key)
				results = append(results, _json)
			}
			key, value = cursor.Next()
		}
		var sorted sort.Interface
		sorted = &Results{results, constraints.order.path}
		if constraints.order.dir == "desc" {
			sorted = sort.Reverse(sorted)
		}
		sort.Sort(sorted)
		if constraints.hasOffset && constraints.offset < len(results) {
			results = results[constraints.offset:]
		}
		if constraints.hasLimit && constraints.limit <= len(results) {
			results = results[:constraints.limit]
		}
	} else {
		cursor := bucket.Cursor()
		key, value := cursor.Next()
		if constraints.hasOffset {
			offset := 0
			for key != nil && offset < constraints.offset {
				key, value = cursor.Next()
			}
		}
		for key != nil {
			_value := make([]byte, len(value))
			copy(_value, value)
			decoder := json.NewDecoder(bytes.NewReader(_value))
			decoder.UseNumber()
			var _json interface{}
			err = decoder.Decode(&_json)
			if err != nil {
				return err, nil
			}
			if process(ast, &Context{buffer, _json, params}).b {
				_key := make([]byte, len(key))
				copy(_key, key)
				_json.(map[string]interface{})["$id"] = string(_key)
				results = append(results, _json)
				if constraints.hasLimit && len(results) == constraints.limit {
					break
				}
			}
			key, value = cursor.Next()
		}
	}

	return nil, results
}

func (s *Store) Shutdown() {
	if s.boltdb != nil {
		s.boltdb.Close()
	}
}

type Value struct {
	b      bool
	errors []error
}

type Constraints struct {
	errors []error
	order  struct {
		path []string
		dir  string
	}
	hasLimit  bool
	limit     int
	hasOffset bool
	offset    int
}

type Context struct {
	buffer []rune
	json   interface{}
	param  []interface{}
}

type Results struct {
	results []interface{}
	path    []string
}

var _ sort.Interface = (*Results)(nil)

func (r *Results) Len() int {
	return len(r.results)
}

func (r *Results) walkPath(i int) interface{} {
	_json, valid := r.results[i], false
	for _, path := range r.path {
		_json, valid = _json.(map[string]interface{})[path]
		if !valid {
			return nil
		}
	}
	return _json
}

func (r *Results) Less(i, j int) bool {
	ii, jj := r.walkPath(i), r.walkPath(j)
	if ii == nil || jj == nil {
		return false
	}
	switch ii := ii.(type) {
	case string:
		if jj, valid := jj.(string); valid {
			return ii < jj
		}
	case json.Number:
		if jj, valid := jj.(json.Number); valid {
			iii, jjj := &big.Rat{}, &big.Rat{}
			iii.SetString(ii.String())
			jjj.SetString(jj.String())
			return iii.Cmp(jjj) < 0
		}
	case bool:
		if jj, valid := jj.(bool); valid {
			return !ii && jj
		}
	}
	return false
}

func (r *Results) Swap(i, j int) {
	r.results[i], r.results[j] = r.results[j], r.results[i]
}

func getConstraints(node *node32, context *Context) (c Constraints) {
	for node != nil {
		switch node.pegRule {
		case rulee:
			return getConstraints(node.up, context)
		case ruleorder:
			c.processOrder(node.up, context)
		case rulelimit:
			c.processLimit(node.up, context)
		case ruleoffset:
			c.processOffset(node.up, context)
		}
		node = node.next
	}
	return
}

func (c *Constraints) processOrder(node *node32, context *Context) {
	for node != nil {
		switch node.pegRule {
		case rulepath:
			path := node.up
			for path != nil {
				if path.pegRule == ruleword {
					c.order.path = append(c.order.path, string(context.buffer[path.begin:path.end]))
				}
				path = path.next
			}
		case ruleasc:
			c.order.dir = string(context.buffer[node.begin:node.end])
		case ruledesc:
			c.order.dir = string(context.buffer[node.begin:node.end])
		}
		node = node.next
	}
}

func (c *Constraints) processLimit(node *node32, context *Context) {
	c.hasLimit = true
	for node != nil {
		if node.pegRule == rulevalue1 {
			var err error
			c.limit, err = processValue1(node.up, context)
			if err != nil {
				c.errors = append(c.errors, err)
			}
		}
		node = node.next
	}
}

func (c *Constraints) processOffset(node *node32, context *Context) {
	c.hasOffset = true
	for node != nil {
		if node.pegRule == rulevalue1 {
			var err error
			c.offset, err = processValue1(node.up, context)
			if err != nil {
				c.errors = append(c.errors, err)
			}
		}
		node = node.next
	}
}

func processValue1(node *node32, context *Context) (int, error) {
	for node != nil {
		switch node.pegRule {
		case ruleplaceholder:
			placeholder, err := strconv.Atoi(string(context.buffer[node.begin+1 : node.end]))
			if err != nil {
				return -1, err
			}
			if placeholder > len(context.param) {
				return -1, errors.New("placholder too large")
			}
			if holder, valid := context.param[placeholder-1].(int); valid {
				return holder, nil
			} else {
				return -1, errors.New("value must be type int")
			}
		case rulewhole:
			whole, err := strconv.Atoi(string(context.buffer[node.begin:node.end]))
			if err != nil {
				return -1, err
			}
			return whole, nil
		}
		node = node.next
	}
	return -1, errors.New("no value")
}

func process(node *node32, context *Context) (v Value) {
	for node != nil {
		switch node.pegRule {
		case rulee:
			return process(node.up, context)
		case rulee1:
			v = processRulee1(node.up, context)
		}
		node = node.next
	}
	return
}

func processRulee1(node *node32, context *Context) (v Value) {
	for node != nil {
		if node.pegRule == rulee2 {
			if !v.b {
				x := processRulee2(node.up, context)
				v.b = v.b || x.b
				v.errors = append(v.errors, x.errors...)
			}
		}
		node = node.next
	}
	return
}

func processRulee2(node *node32, context *Context) (v Value) {
	v.b = true
	for node != nil {
		if node.pegRule == rulee3 {
			if v.b {
				x := processRulee3(node.up, context)
				v.b = v.b && x.b
				v.errors = append(v.errors, x.errors...)
			}
		}
		node = node.next
	}
	return
}

func processRulee3(node *node32, context *Context) (v Value) {
	if node.pegRule == ruleexpression {
		return processExpression(node.up, context)
	}
	return process(node.next.up, context)
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

func processExpression(node *node32, context *Context) (v Value) {
	path, _json, valid := node.up, context.json, false
	for path != nil {
		if path.pegRule == ruleword {
			_json, valid = _json.(map[string]interface{})[string(context.buffer[path.begin:path.end])]
			if !valid {
				return
			}
		}
		path = path.next
	}
	node = node.next
	op := strings.TrimSpace(string(context.buffer[node.begin:node.end]))
	node = node.next.up
	switch node.pegRule {
	case ruleplaceholder:
		placeholder, err := strconv.Atoi(string(context.buffer[node.begin+1 : node.end]))
		if err != nil {
			v.errors = append(v.errors, err)
			return
		}

		if placeholder > len(context.param) {
			v.errors = append(v.errors, errors.New("placholder to large"))
			return
		}
		switch _b := context.param[placeholder-1].(type) {
		case string:
			if _a, valid := _json.(string); valid {
				v.b = compareString(op, _a, _b)
			}
		case float64:
			if _a, valid := _json.(json.Number); valid {
				a, b := &big.Rat{}, &big.Rat{}
				a.SetString(_a.String())
				b.SetFloat64(_b)
				v.b = compareRat(op, a, b)
			}
		case int:
			if _a, valid := _json.(json.Number); valid {
				a, b := &big.Rat{}, &big.Rat{}
				a.SetString(_a.String())
				b.SetInt64(int64(_b))
				v.b = compareRat(op, a, b)
			}
		case bool:
			if a, valid := _json.(bool); valid {
				v.b = compareBool(op, a, _b)
			}
		default:
			v.b = compareNull(op, _json, _b)
		}
	case rulestring:
		b := string(context.buffer[node.begin+1 : node.end-1])
		if a, valid := _json.(string); valid {
			v.b = compareString(op, a, b)
		}
	case rulenumber:
		_b := string(context.buffer[node.begin:node.end])
		b := &big.Rat{}
		b.SetString(_b)
		if _a, valid := _json.(json.Number); valid {
			a := &big.Rat{}
			a.SetString(_a.String())
			v.b = compareRat(op, a, b)
		}
	case ruleboolean:
		b := string(context.buffer[node.begin:node.end])
		if a, valid := _json.(bool); valid {
			v.b = compareBool(op, a, b == "true")
		}
	case rulenull:
		v.b = compareNull(op, _json, nil)
	}
	return
}
