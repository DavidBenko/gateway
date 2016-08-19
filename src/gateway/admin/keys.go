package admin

import (
	"errors"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type KeysController struct {
	BaseController
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
		aphttp.NewError(errors.New("No push channel matches"), 404)
	}

	return serializeCollection(keys, w)
}

func (k *KeysController) Create(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	file, _, err := r.FormFile("key")
	if err != nil {
		return aphttp.NewError(errors.New("invalid file"), 400)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return aphttp.NewError(errors.New("unable to read file"), 400)
	}

	name := r.FormValue("name")

	accountID := k.accountID(r)
	userID := k.userID(r)

	key := &model.Key{Name: name, Key: data, AccountID: accountID}

	if validationErrors := key.Validate(true); !validationErrors.Empty() {
		return SerializableValidationErrors{validationErrors}
	}

	if err = key.Insert(accountID, userID, 0, tx); err != nil {
		validationErrors := key.ValidateFromDatabaseError(err)
		return SerializableValidationErrors{validationErrors}
	}

	return nil
}

func (k *KeysController) Delete(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	keyID := parseKeyID(r)

	key := &model.Key{ID: keyID}
	accountID := k.accountID(r)

	err := key.Delete(accountID, tx)
	if err != nil {
		logreport.Print(err)
		return aphttp.NewError(errors.New("failed to delete key"), 500)
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
