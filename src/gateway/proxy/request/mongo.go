package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/mgo.v2"
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

type operation func(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse)

func operationFind(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse) {
	query := arguments["2"]
	if query == nil {
		query = map[string]interface{}{}
	}
	if _, valid := query.(map[string]interface{}); !valid {
		response.Error = "query parameter is not an object"
		return
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
}

func operationInsert(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse) {
	arg := arguments["2"]
	var err error
	if docs, ok := arg.([]interface{}); ok {
		for _, doc := range docs {
			if _, valid := doc.(map[string]interface{}); !valid {
				response.Error = "document is not an object"
				return
			}
			normalizeObjectId(doc.(map[string]interface{}))
		}
		err = collection.Insert(arg.([]interface{})...)
	} else {
		if _, valid := arg.(map[string]interface{}); !valid {
			response.Error = "document is not an object"
			return
		}
		normalizeObjectId(arg.(map[string]interface{}))
		err = collection.Insert(arg)
	}
	if err != nil {
		response.Error = err.Error()
	}
}

func operationUpdate(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse) {
	query := arguments["2"]
	if _, valid := query.(map[string]interface{}); !valid {
		response.Error = "query parameter is not an object"
		return
	}
	normalizeObjectId(query.(map[string]interface{}))
	update := arguments["3"]
	if _, valid := update.(map[string]interface{}); !valid {
		response.Error = "update parameter is not an object"
		return
	}
	normalizeObjectId(update.(map[string]interface{}))

	upsert, multi := false, false
	if len(arguments) > 4 {
		_options := arguments["4"]
		if _, valid := _options.(map[string]interface{}); !valid {
			response.Error = "options parameter should be an object"
			return
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
				return
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
				return
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
}

func operationSave(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse) {
	d := arguments["2"]
	if _, valid := d.(map[string]interface{}); !valid {
		response.Error = "save requires a document to save"
		return
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
}

func operationRemove(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse) {
	query := arguments["2"]
	if query == nil {
		query = map[string]interface{}{}
	}
	if _, valid := query.(map[string]interface{}); !valid {
		response.Error = "query parameter is not an object"
		return
	}
	normalizeObjectId(query.(map[string]interface{}))

	justOne := false
	if len(arguments) > 3 {
		_justOne := arguments["3"]
		switch _justOne := _justOne.(type) {
		case int64:
			justOne = _justOne == 1
		case bool:
			justOne = _justOne
		default:
			response.Error = "just one should be a boolean value"
			return
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
}

func operationDrop(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse) {
	err := collection.DropCollection()
	if err != nil {
		response.Error = err.Error()
	}
}

func operationAggregate(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse) {
	length := len(arguments)
	if length < 3 {
		response.Error = "aggregate requires at least one pipline stage"
		return
	}

	var stages []map[string]interface{}
	if arg, valid := arguments["2"].([]interface{}); valid {
		stages = make([]map[string]interface{}, len(arg))
		for i, stage := range arg {
			if stage, valid := stage.(map[string]interface{}); valid {
				stages[i] = stage
			} else {
				response.Error = "stage is not an object"
				return
			}
		}
	} else {
		stages = make([]map[string]interface{}, length-2)
		for i := 2; i < length; i++ {
			if stage, valid := arguments[fmt.Sprintf("%v", i)].(map[string]interface{}); valid {
				stages[i-2] = stage
			} else {
				response.Error = "stage is not an object"
				return
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
		return
	}
	response.Count = len(response.Data)
}

func operationCount(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse) {
	query := arguments["2"]
	if query == nil {
		query = map[string]interface{}{}
	}
	if _, valid := query.(map[string]interface{}); !valid {
		response.Error = "query parameter is not an object"
		return
	}
	normalizeObjectId(query.(map[string]interface{}))
	n, err := collection.Find(query).Count()
	if err != nil {
		response.Error = err.Error()
		return
	}
	response.Count = n
}

func operationMapReduce(arguments map[string]interface{},
	collection *mgo.Collection, response *MongoResponse) {
	query := arguments["2"]
	if _, valid := query.(map[string]interface{}); !valid {
		response.Error = "query parameter is not an object"
		return
	}
	normalizeObjectId(query.(map[string]interface{}))
	mr := mgo.MapReduce{}
	if scope, valid := arguments["3"].(map[string]interface{}); valid {
		mr.Scope = scope
	} else {
		response.Error = "scope parameter is not an object"
		return
	}
	if _map, valid := arguments["4"].(string); valid {
		mr.Map = _map
	} else {
		response.Error = "map parameter is not a string"
		return
	}
	if reduce, valid := arguments["5"].(string); valid {
		mr.Reduce = reduce
	} else {
		response.Error = "reduce parameter is not a string"
		return
	}
	if finalize, valid := arguments["6"].(string); valid {
		mr.Finalize = finalize
	} else {
		response.Error = "finalize parameter is not a string"
		return
	}
	_query := collection.Find(query)
	if out, valid := arguments["7"].(map[string]interface{}); valid {
		mr.Out = out
		_, err := _query.MapReduce(&mr, nil)
		if err != nil {
			response.Error = err.Error()
			return
		}
	} else {
		response.Data = []map[string]interface{}{}
		_, err := _query.MapReduce(&mr, &response.Data)
		if err != nil {
			response.Error = err.Error()
			return
		}
		response.Count = len(response.Data)
	}
}

var operations = map[string]operation{
	"find":      operationFind,
	"insert":    operationInsert,
	"update":    operationUpdate,
	"save":      operationSave,
	"remove":    operationRemove,
	"drop":      operationDrop,
	"aggregate": operationAggregate,
	"count":     operationCount,
	"mapReduce": operationMapReduce,
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
	if op, valid := operations[op.(string)]; valid {
		op(r.Arguments, collection, response)
	} else {
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
