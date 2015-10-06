package remote_endpoint

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx/types"
)

// Soap represents a configuration for a remote soap endpoint
type Soap struct {
	WSDL string `json:"wsdl"`
}

// SoapConfig creates a new Soap object from a JsonText object
func SoapConfig(data types.JsonText) (*Soap, error) {
	conf := &Soap{}

	err := json.Unmarshal(data, conf)
	if err != nil {
		return nil, fmt.Errorf("bad JSON for Soap config: %s", err.Error())
	}

	return conf, nil
}
