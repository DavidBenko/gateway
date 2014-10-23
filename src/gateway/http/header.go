package http

import "net/http"

// AddHeaders adds the headers in headers to the Header! Yo dawg.
func AddHeaders(h http.Header, headers map[string]interface{}) {
	for key, value := range headers {
		switch value := value.(type) {
		case string:
			h.Add(key, value)
		case []interface{}:
			for _, v := range value {
				AddHeaders(h, map[string]interface{}{key: v})
			}
		}
	}
}

// DesliceValues is used to collapse single value string slices from map values.
func DesliceValues(slice map[string][]string) map[string]interface{} {
	desliced := make(map[string]interface{})
	for k, v := range slice {
		if len(v) == 1 {
			desliced[k] = v[0]
		} else {
			desliced[k] = v
		}
	}
	return desliced
}

// ResliceValues is the opposite of DesliceValues, and turns single value
// strings into slices of string in map values.
func ResliceValues(slice map[string]string) map[string][]string {
	resliced := make(map[string][]string)
	for k, v := range slice {
		resliced[k] = []string{v}
	}
	return resliced
}
