package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"gateway/model"
	"gateway/store"
)

type StoreRequest struct {
	Arguments map[string]interface{} `json:"arguments"`
	Store     store.Store            `json:"-"`
	AccountID int64                  `json:"-"`
}

type storeOperation func(request *StoreRequest) ([]map[string]interface{}, error)

func storeOperationInsert(request *StoreRequest) ([]map[string]interface{}, error) {
	var data []map[string]interface{}

	collection, valid := request.Arguments["1"].(string)
	if !valid {
		return nil, errors.New("collection is not a string")
	}
	object := request.Arguments["2"]
	_, isObject := object.(interface{})
	_, isArray := object.([]interface{})
	if !(isObject || isArray) {
		return nil, errors.New("object is not an Object or Array")
	}
	results, err := request.Store.Insert(request.AccountID, collection, object)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		data = append(data, result.(map[string]interface{}))
	}

	return data, nil
}

type Argument struct {
	key   string
	valud interface{}
}

func storeOperationSelect(request *StoreRequest) ([]map[string]interface{}, error) {
	var data []map[string]interface{}

	collection, valid := request.Arguments["1"].(string)
	if !valid {
		return nil, errors.New("collection is not a string")
	}
	switch query := request.Arguments["2"].(type) {
	case string:
		arguments := make([]interface{}, len(request.Arguments))
		for key, value := range request.Arguments {
			i, err := strconv.Atoi(key)
			if err != nil {
				return nil, err
			}
			arguments[i] = value
		}
		results, err := request.Store.Select(request.AccountID, collection, query, arguments[3:]...)
		if err != nil {
			return nil, err
		}
		for _, result := range results {
			data = append(data, result.(map[string]interface{}))
		}
	case float64:
		result, err := request.Store.SelectByID(request.AccountID, collection, uint64(query))
		if err != nil {
			return nil, err
		}
		data = append(data, result.(map[string]interface{}))
	default:
		return nil, errors.New("invalid type for query")
	}

	return data, nil
}

func storeOperationUpdate(request *StoreRequest) ([]map[string]interface{}, error) {
	var data []map[string]interface{}

	collection, valid := request.Arguments["1"].(string)
	if !valid {
		return nil, errors.New("collection is not a string")
	}
	id, valid := request.Arguments["2"].(float64)
	if !valid {
		return nil, errors.New("id is not a number")
	}
	object, valid := request.Arguments["3"].(map[string]interface{})
	if !valid {
		return nil, errors.New("object is not an Object")
	}
	result, err := request.Store.UpdateByID(request.AccountID, collection, uint64(id), object)
	if err != nil {
		return nil, err
	}
	data = append(data, result.(map[string]interface{}))

	return data, nil
}

func storeOperationDelete(request *StoreRequest) ([]map[string]interface{}, error) {
	var data []map[string]interface{}

	collection, valid := request.Arguments["1"].(string)
	if !valid {
		return nil, errors.New("collection is not a string")
	}
	switch query := request.Arguments["2"].(type) {
	case string:
		arguments := make([]interface{}, len(request.Arguments))
		for key, value := range request.Arguments {
			i, err := strconv.Atoi(key)
			if err != nil {
				return nil, err
			}
			arguments[i] = value
		}
		results, err := request.Store.Delete(request.AccountID, collection, query, arguments[3:]...)
		if err != nil {
			return nil, err
		}
		for _, result := range results {
			data = append(data, result.(map[string]interface{}))
		}
	case float64:
		result, err := request.Store.DeleteByID(request.AccountID, collection, uint64(query))
		if err != nil {
			return nil, err
		}
		data = append(data, result.(map[string]interface{}))
	default:
		return nil, errors.New("invalid type for query")
	}

	return data, nil
}

var storeOperations = map[string]storeOperation{
	"insert": storeOperationInsert,
	"select": storeOperationSelect,
	"update": storeOperationUpdate,
	"delete": storeOperationDelete,
}

func (r *StoreRequest) Perform() (_response Response) {
	response := &StoreResponse{}
	_response = response
	defer func() {
		if r := recover(); r != nil {
			response.Error = fmt.Sprintf("%v", r)
		}
	}()

	op := r.Arguments["0"]
	if _, valid := op.(string); !valid {
		response.Error = "Missing operation parameter"
		return
	}
	if op, valid := storeOperations[op.(string)]; valid {
		data, err := op(r)
		if err != nil {
			response.Error = err.Error()
		}
		response.Data = data
	} else {
		response.Error = "Invalid operation"
	}

	return
}

func (request *StoreRequest) Log(devMode bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("\nArguments: %s", request.Arguments))
	if devMode {
		buffer.WriteString(fmt.Sprintf("\nAccountID: %d", request.AccountID))
	}
	return buffer.String()
}

func (request *StoreRequest) JSON() ([]byte, error) {
	return json.Marshal(request)
}

type StoreResponse struct {
	Data  []map[string]interface{} `json:"data"`
	Error string                   `json:"error,omitempty"`
}

func (r *StoreResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

func (r *StoreResponse) Log() string {
	if r.Data != nil {
		return fmt.Sprintf("Records found: %d", len(r.Data))
	}

	return r.Error
}

func NewStoreRequest(s store.Store, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &StoreRequest{Store: s, AccountID: endpoint.AccountID}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	return request, nil
}
