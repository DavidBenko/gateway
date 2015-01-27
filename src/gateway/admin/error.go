package admin

import (
	"encoding/json"
	"fmt"
	"gateway/model"
)

type SerializableValidationErrors struct {
	Errors model.Errors `json:"errors"`
}

func (e SerializableValidationErrors) Error() error {
	return nil
}

func (e SerializableValidationErrors) Body() string {
	errorsJSON, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		return fmt.Sprintf("%s", e.Errors)
	}
	return string(errorsJSON)
}

func (e SerializableValidationErrors) String() string {
	return e.Body()
}

func (e SerializableValidationErrors) Code() int {
	return 400
}
