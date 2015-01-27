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
	return parseID(mux.Vars(r)["id"])
}

func accountIDFromPath(r *http.Request) int64 {
	return parseID(mux.Vars(r)["accountID"])
}

func accountIDFromSession(r *http.Request) int64 {
	session := requestSession(r)
	return session.Values[accountIDKey].(int64)
}

func apiIDFromPath(r *http.Request) int64 {
	return parseID(mux.Vars(r)["apiID"])
}

func parseID(id string) int64 {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return -1
	}
	return i
}

func deserialize(dest interface{}, r *http.Request) aphttp.Error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return aphttp.NewError(err, http.StatusInternalServerError)
	}

	err = json.Unmarshal(body, dest)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
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

// To be removed when SerializedValidationErrors are adopted everywhere
type wrappedErrors struct {
	Errors model.Errors `json:"errors"`
}
