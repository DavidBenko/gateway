package proxy

import (
  "bytes"
  "encoding/json"
  "fmt"

  "gopkg.in/mgo.v2/bson"

  "gateway/db/mongo"
)

type MongoRequest struct {
  Arguments map[string] interface{} `json:"arguments"`
  Config    mongo.Conn              `json:"config"`
  Limit     int                     `json:"limit"`
  conn      *mongo.DB
}

func normalizeObjectId(m map[string] interface{}) {
  for key, value := range m {
    switch value := value.(type) {
    case map[string] interface{}:
      if id := value["_id"]; id != nil && len(value) == 1 {
        m[key] = bson.ObjectIdHex(id.(string))
      } else {
        normalizeObjectId(value)
      }
    case []interface{}:
      for _, value := range value {
        if value, valid := value.(map[string] interface{}); valid {
          normalizeObjectId(value)
        }
      }
    }
  }
}

func (r *MongoRequest) Perform() Response {
  response := &MongoResponse{Type: "mongodb"}
  c := r.Arguments["0"]
  if c == nil {
    response.Error = "Missing collection parameter"
    return response
  }
  collection := r.conn.DB(r.Config["database"].(string)).C(c.(string))
  op := r.Arguments["1"]
  if op == nil {
    response.Error = "Missing operation parameter"
    return response
  }
  switch op.(string) {
  case "find":
    query := r.Arguments["2"]
    if query == nil {
      query = map[string] interface{}{}
    }
    normalizeObjectId(query.(map[string] interface{}))
    iter := collection.Find(query).Iter()
    record := bson.M{}
    for iter.Next(&record) {
      if id, valid := record["_id"].(bson.ObjectId); valid {
        record["_id"] = map[string] interface{}{"_id": id.Hex()}
      }
      response.Data, record = append(response.Data, record), bson.M{}
    }
  case "insert":
    arg := r.Arguments["2"]
    if arg == nil {
      response.Error = "insert requires a document or an array of documents to insert"
      break
    }
    var err error
    if docs, ok := arg.([]interface{}); ok {
      for _, doc := range docs {
        normalizeObjectId(doc.(map[string] interface{}))
      }
      err = collection.Insert(arg.([]interface{})...)
    } else {
      normalizeObjectId(arg.(map[string] interface{}))
      err = collection.Insert(arg)
    }
    if err != nil {
      response.Error = err.Error()
    }
  case "update":
    query := r.Arguments["2"]
    if query == nil {
      response.Error = "update requires a query parameter"
      break
    }
    normalizeObjectId(query.(map[string] interface{}))
    update := r.Arguments["3"]
    if update == nil {
      response.Error = "update requires an update parameter"
      break
    }
    normalizeObjectId(update.(map[string] interface{}))

    upsert, multi := false, false
    if len(r.Arguments) > 4 {
      options := r.Arguments["4"].(map[string] interface{})
      if _upsert := options["upsert"]; _upsert != nil {
        switch _upsert := _upsert.(type) {
        case int64:
          upsert = _upsert == 1
        case bool:
          upsert = _upsert
        }
      }
      if _multi := options["multi"]; _multi != nil {
        switch _multi := _multi.(type) {
        case int64:
          multi = _multi == 1
        case bool:
          multi = _multi
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
    if d == nil {
      response.Error = "save requires a document to save"
      break
    }
    doc := d.(map[string] interface{})
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
      query = map[string] interface{}{}
    }
    normalizeObjectId(query.(map[string] interface{}))

    justOne := false
    if len(r.Arguments) > 3 {
      _justOne := r.Arguments["3"]
      switch _justOne := _justOne.(type) {
      case int64:
        justOne = _justOne == 1
      case bool:
        justOne = _justOne
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
