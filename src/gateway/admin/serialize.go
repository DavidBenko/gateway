package admin

import (
	"encoding/json"
	"fmt"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func instanceID(r *http.Request) int64 {
	id := mux.Vars(r)["id"]
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return -1
	}
	return i
}

func deserialize(dest interface{}, r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("%s Error reading body: %v", config.System, err)
		return err
	}

	err = json.Unmarshal(body, dest)
	if err != nil {
		log.Printf("%s Error deserializing data: %v, %v", config.System, err, body)
		return err
	}

	return nil
}

func serialize(data interface{}, w http.ResponseWriter) aphttp.Error {
	dataJSON, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("%s Error serializing data: %v, %v", config.System, err, data)
		return aphttp.DefaultServerError()
	}

	fmt.Fprintf(w, "%s\n", string(dataJSON))
	return nil
}

type wrappedErrors struct {
	Errors model.Errors `json:"errors"`
}
