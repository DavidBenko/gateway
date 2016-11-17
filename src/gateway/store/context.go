package store

import (
	"encoding/json"
	"strconv"
	"strings"
)

type context struct {
	buffer []rune
}

func (c *context) Node(node *node32) string {
	return strings.TrimSpace(string(c.buffer[node.begin:node.end]))
}

func (c *context) getFloat64(path *node32, _json interface{}) (value float64, valid bool) {
	for path != nil {
		if path.pegRule == ruleword {
			_json, valid = _json.(map[string]interface{})[c.Node(path)]
			if !valid {
				return 0, valid
			}
		}
		path = path.next
	}
	switch value := _json.(type) {
	case json.Number:
		v, err := value.Float64()
		if err != nil {
			return 0, false
		}
		return v, true
	case float64:
		return value, true
	case string:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, false
		}
		return v, true
	case bool:
		if value {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func (c *context) getValid(path *node32, _json interface{}) (valid bool) {
	for path != nil {
		if path.pegRule == ruleword {
			_json, valid = _json.(map[string]interface{})[c.Node(path)]
			if !valid {
				return valid
			}
		}
		path = path.next
	}
	return true
}
