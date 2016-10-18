package vm

import (
	"encoding/json"
)

// PrepareRawRequests runs AP.prepareRequests on a slice of named calls to
// generate a slice of *json.RawMessage, each corresponding to a
// 'gateway/proxy/request' Request.
func (c *CoreVM) PrepareRawRequests(calls []string) ([]*json.RawMessage, error) {
	obj, err := c.makeCall(prepareRequests, calls)
	if err != nil {
		return nil, err
	}
	requestsJSON := obj.String()

	var rawRequests []*json.RawMessage
	err = json.Unmarshal([]byte(requestsJSON), &rawRequests)
	if err != nil {
		return nil, err
	}

	return rawRequests, nil
}
