package request

import (
	"fmt"
	"strconv"
)

var converters = map[string]map[string]conversion{
	"int64": {
		"float64": toFloat64,
		"int64":   toInt64,
		"bool":    toBool,
		"string":  toString,
	},
	"int32": {
		"float64": toFloat64,
		"int64":   toInt64,
		"bool":    toBool,
		"string":  toString,
	},
	"int16": {
		"float64": toFloat64,
		"int64":   toInt64,
		"bool":    toBool,
		"string":  toString,
	},
	"int8": {
		"float64": toFloat64,
		"int64":   toInt64,
		"bool":    toBool,
		"string":  toString,
	},
	"float64": {
		"int64":   toInt64,
		"float64": toFloat64,
		"bool":    toBool,
		"string":  toString,
	},
	"float32": {
		"int64":   toInt64,
		"float64": toFloat64,
		"bool":    toBool,
		"string":  toString,
	},
	"uint64": {
		"int64":   toInt64,
		"float64": toFloat64,
		"bool":    toBool,
		"string":  toString,
	},
	"uint32": {
		"int64":   toInt64,
		"float64": toFloat64,
		"bool":    toBool,
		"string":  toString,
	},
	"uint16": {
		"int64":   toInt64,
		"float64": toFloat64,
		"bool":    toBool,
		"string":  toString,
	},
	"uint8": {
		"int64":   toInt64,
		"float64": toFloat64,
		"bool":    toBool,
		"string":  toString,
	},
	"bool": {
		"int64":   toInt64,
		"float64": toFloat64,
		"bool":    toBool,
		"string":  toString,
	},
	"string": {
		"int64":   toInt64,
		"float64": toFloat64,
		"bool":    toBool,
		"string":  toString,
	},
}

type conversion func(interface{}) (interface{}, error)

func toString(src interface{}) (interface{}, error) {
	switch s := src.(type) {
	case float64, float32, uint64, uint32, uint16, uint8, int64, int32, int16, int8:
		return fmt.Sprintf("%d", s), nil
	case string:
		return s, nil
	case bool:
		return fmt.Sprintf("%t", s), nil
	default:
		return nil, fmt.Errorf("Unable to convert from %T to string", s)
	}
}

func toBool(src interface{}) (interface{}, error) {
	switch s := src.(type) {
	case float64, float32:
		return s == 1.0, nil
	case uint64, uint32, uint16, uint8, int64, int32, int16, int8:
		return s == 1, nil
	case string:
		return strconv.ParseBool(s)
	case bool:
		return s, nil
	default:
		return nil, fmt.Errorf("Unable to convert from %T to bool", s)
	}
}

func toFloat64(src interface{}) (interface{}, error) {
	switch s := src.(type) {
	case float64:
		return s, nil
	case float32:
		return float64(s), nil
	case uint64:
		return float64(s), nil
	case uint32:
		return float64(s), nil
	case uint16:
		return float64(s), nil
	case uint8:
		return float64(s), nil
	case int64:
		return float64(s), nil
	case int32:
		return float64(s), nil
	case int16:
		return float64(s), nil
	case int8:
		return float64(s), nil
	case string:
		return strconv.ParseFloat(s, 64)
	case bool:
		if s {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return nil, fmt.Errorf("Unable to convert from %T to float64", s)
	}
}

func toInt64(src interface{}) (interface{}, error) {
	switch s := src.(type) {
	case int64:
		return s, nil
	case int32:
		return int64(s), nil
	case int16:
		return int64(s), nil
	case int8:
		return int64(s), nil
	case float64:
		return int64(s), nil
	case float32:
		return int64(s), nil
	case uint64:
		return int64(s), nil
	case uint32:
		return int64(s), nil
	case uint16:
		return int64(s), nil
	case uint8:
		return int64(s), nil
	case string:
		return strconv.ParseInt(s, 0, 64)
	case bool:
		if s {
			return 1, nil
		}
		return 0, nil
	default:
		return nil, fmt.Errorf("Unable to convert from %T to int64", s)
	}
}
