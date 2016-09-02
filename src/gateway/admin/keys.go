package admin

import (
	"errors"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/vincent-petithory/dataurl"
)

type KeysController struct {
	BaseController
}

func deserializeInstance(file io.Reader) (*model.Key, aphttp.Error) {
	type payloadKey struct {
		Key  string
		Name string
	}

	type wrapped struct {
		Key *payloadKey
	}

	w := &wrapped{}
	if err := deserialize(&w, file); err != nil {
		return nil, err
	}
	if w.Key == nil {
		return nil, aphttp.NewError(errors.New("Could not deserialize key from JSON"), http.StatusBadRequest)
	}

	data, err := dataurl.DecodeString(w.Key.Key)
	if err != nil {
		logreport.Printf("%s Error getting form file: %v\n", config.Admin, err)
		return nil, aphttp.NewError(errors.New("invalid file"), http.StatusBadRequest)
	}

	key := &model.Key{Name: w.Key.Name, Key: data.Data}

	return key, nil
}

func RouteKeys(controller *KeysController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET":  read(db, controller.List),
		"POST": write(db, controller.Create),
	}
	instanceRoutes := map[string]http.Handler{
		"DELETE": write(db, controller.Delete),
	}

	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "POST", "OPTIONS"})
		instanceRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"DELETE"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
	router.Handle(path+"/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler(instanceRoutes)))
}

func (k *KeysController) List(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
	accountID := k.accountID(r)
	keys, err := model.FindKeysForAccount(accountID, db)
	if err != nil {
		logreport.Printf("%s Error listing keys: %v\n%v", config.Admin, err, r)
		aphttp.NewError(errors.New("no keys found"), http.StatusNotFound)
	}

	return serializeCollection(keys, w)
}

func (k *KeysController) Create(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	key, err := deserializeInstance(r.Body)
	if err != nil {
		return err
	}

	accountID := k.accountID(r)
	userID := k.userID(r)

	key.AccountID = accountID

	if validationErrors := key.Validate(true); !validationErrors.Empty() {
		return SerializableValidationErrors{validationErrors}
	}

	if err := key.Insert(accountID, userID, 0, tx); err != nil {
		validationErrors := key.ValidateFromDatabaseError(err)
		return SerializableValidationErrors{validationErrors}
	}

	return nil
}

func (k *KeysController) Delete(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	keyID := parseKeyID(r)

	key := &model.Key{ID: keyID}
	accountID := k.accountID(r)
	userID := k.userID(r)

	err := key.Delete(accountID, userID, 0, tx)
	if err != nil {
		logreport.Printf("%s Error deleting key: %v\n%v", config.Admin, err, r)
		return aphttp.NewError(errors.New("failed to delete key"), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func parseKeyID(r *http.Request) int64 {
	s := mux.Vars(r)["id"]
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}
	return v
}

func serializeCollection(collection []*model.Key, w http.ResponseWriter) aphttp.Error {
	wrapped := struct {
		Keys []*model.Key `json:"keys"`
	}{collection}
	return serialize(wrapped, w)
}
