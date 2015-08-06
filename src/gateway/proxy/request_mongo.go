package proxy

import (
  "bytes"
  "encoding/json"
  "fmt"

  "gateway/db/mongo"
)

type MongoRequest struct {
  Config  mongo.Conn `json:"config"`
  Limit   int        `json:"limit"`
  conn    *mongo.DB
}

func (r *MongoRequest) Perform() Response {
  return &MongoResponse{}
}

func (request *MongoRequest) Log(devMode bool) string {
	var buffer bytes.Buffer

	if devMode {
		buffer.WriteString(fmt.Sprintf("\nConnection: %s", request.Config))
	}
	return buffer.String()
}

type MongoResponse struct {

}

func (r *MongoResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

func (r *MongoResponse) Log() string {
	return "MongoResponse"
}
