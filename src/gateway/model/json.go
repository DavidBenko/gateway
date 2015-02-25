package model

import "github.com/jmoiron/sqlx/types"

func marshaledForStorage(json types.JsonText) (string, error) {
	data, err := json.MarshalJSON()
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "null", nil
	}
	return string(data), nil
}
