package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"gateway/db/mongo"
	"gateway/db/pools"
	"gateway/model"
)

type MongoRequest struct {
	Arguments map[string]interface{} `json:"arguments"`
	Config    mongo.Conn             `json:"config"`
	Limit     int                    `json:"limit"`
	conn      *mongo.DB
}

func normalizeObjectId(m map[string]interface{}) {
	for key, value := range m {
		switch value := value.(type) {
		case map[string]interface{}:
			if id, _type := value["_id"], value["type"]; id != nil && _type == "id" && len(value) == 2 {
				m[key] = bson.ObjectIdHex(id.(string))
			} else {
				normalizeObjectId(value)
			}
		case []interface{}:
			for _, value := range value {
				if value, valid := value.(map[string]interface{}); valid {
					normalizeObjectId(value)
				}
			}
		}
	}
}

func (r *MongoRequest) Perform() Response {
	response := &MongoResponse{Type: "mongodb"}
	c := r.Arguments["0"]
	if _, valid := c.(string); !valid {
		response.Error = "Missing collection parameter"
		return response
	}
	collection := r.conn.DB(r.Config["database"].(string)).C(c.(string))
	op := r.Arguments["1"]
	if _, valid := op.(string); !valid {
		response.Error = "Missing operation parameter"
		return response
	}
Operation:
	switch op.(string) {
	case "find":
		query := r.Arguments["2"]
		if query == nil {
			query = map[string]interface{}{}
		}
		if _, valid := query.(map[string]interface{}); !valid {
			response.Error = "query parameter is not an object"
			break
		}
		normalizeObjectId(query.(map[string]interface{}))
		iter := collection.Find(query).Iter()
		record := bson.M{}
		for iter.Next(&record) {
			if id, valid := record["_id"].(bson.ObjectId); valid {
				record["_id"] = map[string]interface{}{"_id": id.Hex(), "type": "id"}
			}
			response.Data, record = append(response.Data, record), bson.M{}
			response.Count++
		}
	case "insert":
		arg := r.Arguments["2"]
		var err error
		if docs, ok := arg.([]interface{}); ok {
			for _, doc := range docs {
				if _, valid := doc.(map[string]interface{}); !valid {
					response.Error = "document is not an object"
					break Operation
				}
				normalizeObjectId(doc.(map[string]interface{}))
			}
			err = collection.Insert(arg.([]interface{})...)
		} else {
			if _, valid := arg.(map[string]interface{}); !valid {
				response.Error = "document is not an object"
				break
			}
			normalizeObjectId(arg.(map[string]interface{}))
			err = collection.Insert(arg)
		}
		if err != nil {
			response.Error = err.Error()
		}
	case "update":
		query := r.Arguments["2"]
		if _, valid := query.(map[string]interface{}); !valid {
			response.Error = "query parameter is not an object"
			break
		}
		normalizeObjectId(query.(map[string]interface{}))
		update := r.Arguments["3"]
		if _, valid := update.(map[string]interface{}); !valid {
			response.Error = "update parameter is not an object"
			break
		}
		normalizeObjectId(update.(map[string]interface{}))

		upsert, multi := false, false
		if len(r.Arguments) > 4 {
			_options := r.Arguments["4"]
			if _, valid := _options.(map[string]interface{}); !valid {
				response.Error = "options parameter should be an object"
				break
			}
			options := _options.(map[string]interface{})
			if _upsert := options["upsert"]; _upsert != nil {
				switch _upsert := _upsert.(type) {
				case int64:
					upsert = _upsert == 1
				case bool:
					upsert = _upsert
				default:
					response.Error = "upsert should be a boolean value"
					break Operation
				}
			}
			if _multi := options["multi"]; _multi != nil {
				switch _multi := _multi.(type) {
				case int64:
					multi = _multi == 1
				case bool:
					multi = _multi
				default:
					response.Error = "multi should be a boolean value"
					break Operation
				}
			}
		}

		var err error
		if upsert {
			_, err = collection.Upsert(query, update)
		} else if multi {
			_, err = collection.UpdateAll(query, update)
		} else {
			err = collection.Update(query, update)
		}
		if err != nil {
			response.Error = err.Error()
		}
	case "save":
		d := r.Arguments["2"]
		if _, valid := d.(map[string]interface{}); !valid {
			response.Error = "save requires a document to save"
			break
		}
		doc := d.(map[string]interface{})
		normalizeObjectId(doc)
		var err error
		if id := doc["_id"]; id != nil {
			_, err = collection.UpsertId(id, doc)
		} else {
			err = collection.Insert(doc)
		}
		if err != nil {
			response.Error = err.Error()
		}
	case "remove":
		query := r.Arguments["2"]
		if query == nil {
			query = map[string]interface{}{}
		}
		if _, valid := query.(map[string]interface{}); !valid {
			response.Error = "query parameter is not an object"
			break
		}
		normalizeObjectId(query.(map[string]interface{}))

		justOne := false
		if len(r.Arguments) > 3 {
			_justOne := r.Arguments["3"]
			switch _justOne := _justOne.(type) {
			case int64:
				justOne = _justOne == 1
			case bool:
				justOne = _justOne
			default:
				response.Error = "just one should be a boolean value"
				break Operation
			}
		}

		var err error
		if justOne {
			err = collection.Remove(query)
		} else {
			_, err = collection.RemoveAll(query)
		}
		if err != nil {
			response.Error = err.Error()
		}
	case "drop":
		err := collection.DropCollection()
		if err != nil {
			response.Error = err.Error()
		}
	case "aggregate":
		length := len(r.Arguments)
		if length < 3 {
			response.Error = "aggregate requires at least one pipline stage"
			break
		}

		var stages []map[string]interface{}
		if arg, valid := r.Arguments["2"].([]interface{}); valid {
			stages = make([]map[string]interface{}, len(arg))
			for i, stage := range arg {
				if stage, valid := stage.(map[string]interface{}); valid {
					stages[i] = stage
				} else {
					response.Error = "stage is not an object"
					break Operation
				}
			}
		} else {
			stages = make([]map[string]interface{}, length-2)
			for i := 2; i < length; i++ {
				if stage, valid := r.Arguments[fmt.Sprintf("%v", i)].(map[string]interface{}); valid {
					stages[i-2] = stage
				} else {
					response.Error = "stage is not an object"
					break Operation
				}
			}
		}
		for _, stage := range stages {
			normalizeObjectId(stage)
		}
		response.Data = []map[string]interface{}{}
		err := collection.Pipe(stages).All(&response.Data)
		if err != nil {
			response.Error = err.Error()
			break
		}
		response.Count = len(response.Data)
	case "count":
		query := r.Arguments["2"]
		if query == nil {
			query = map[string]interface{}{}
		}
		if _, valid := query.(map[string]interface{}); !valid {
			response.Error = "query parameter is not an object"
			break
		}
		normalizeObjectId(query.(map[string]interface{}))
		n, err := collection.Find(query).Count()
		if err != nil {
			response.Error = err.Error()
			break
		}
		response.Count = n
	default:
		response.Error = "Invalid operation"
	}
	return response
}

func (request *MongoRequest) Log(devMode bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("\nArguments: %s", request.Arguments))
	if devMode {
		buffer.WriteString(fmt.Sprintf("\nConnection: %s", request.Config))
	}
	return buffer.String()
}

type MongoResponse struct {
	Type  string                   `json:"type"`
	Data  []map[string]interface{} `json:"data"`
	Count int                      `json:"count"`
	Error string                   `json:"error,omitempty"`
}

func (r *MongoResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

func (r *MongoResponse) Log() string {
	if r.Data != nil {
		return fmt.Sprintf("Records found: %d", len(r.Data))
	}

	return r.Error
}

func NewMongoRequest(pools *pools.Pools, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &MongoRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	endpointData := &MongoRequest{}
	if err := json.Unmarshal(endpoint.Data, endpointData); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal endpoint configuration: %v", err)
	}
	request.updateWith(endpointData)

	if endpoint.SelectedEnvironmentData != nil {
		if err := json.Unmarshal(*endpoint.SelectedEnvironmentData, endpointData); err != nil {
			return nil, err
		}
		request.updateWith(endpointData)
	}

	if pools == nil {
		return nil, errors.New("database pools not set up")
	}

	conn, err := pools.Connect(mongo.Config(
		mongo.Connection(request.Config),
		mongo.PoolLimit(request.Limit),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if mongoConn, ok := conn.(*mongo.DB); ok {
		request.conn = mongoConn
		return request, nil
	}

	return nil, fmt.Errorf("need Mongo connection, got %T", conn)
}

func (r *MongoRequest) updateWith(endpointData *MongoRequest) {
	if endpointData.Arguments != nil {
		r.Arguments = endpointData.Arguments
	}

	if endpointData.Config != nil {
		if r.Config == nil {
			r.Config = mongo.Conn{}
		}
		for key, value := range endpointData.Config {
			r.Config[key] = value
		}
	}

	if r.Limit != endpointData.Limit {
		r.Limit = endpointData.Limit
	}
}
