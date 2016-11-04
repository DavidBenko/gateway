package admin

import (
	"encoding/pem"
	"errors"
	"fmt"
	"gateway/config"
	aperrors "gateway/errors"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"
	"io"
	"net/http"
	"strconv"

	"golang.org/x/crypto/pkcs12"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/vincent-petithory/dataurl"
)

type KeysController struct {
	BaseController
}

func deserializeInstance(file io.Reader) (*model.Key, aphttp.Error) {
	type payloadKey struct {
		Key      string
		Name     string
		Password string
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

	mime := fmt.Sprintf("%s/%s", data.MediaType.Type, data.MediaType.Subtype)
	key := &model.Key{Name: w.Key.Name, Key: data.Data, Mime: mime, Password: w.Key.Password}

	// If the key is a pkcs12 file then parse out the private key using the supplied password.
	// Update the payload's Key to the pem encoded result.
	if key.Mime == "application/x-pkcs12" {
		block, err := parsePkcs12(data.Data, w.Key.Password)
		if err != nil {
			return nil, aphttp.NewError(err, http.StatusBadRequest)
		}
		if block == nil {
			return nil, aphttp.NewError(errors.New("could not find private key"), http.StatusBadRequest)
		}
		key.Key = pem.EncodeToMemory(block)
	}

	return key, nil
}

func parsePkcs12(data []byte, password string) (*pem.Block, error) {
	blocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, errors.New("could not find a valid key")
	}

	// PKCS12 contains a private key and a number of related certificates.
	// Iterate across the blocks until we find the PRIVATE KEY and ignore
	// the rest.
	var block *pem.Block
	for _, b := range blocks {
		if b.Type == "PRIVATE KEY" {
			block = b
			break
		}
	}

	return block, nil
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
		// Hacky way to return a validation error if PKCS12 file failed to parse
		if err.Error().Error() == "could not find a valid key" {
			errors := make(aperrors.Errors)
			errors.Add("key", "could not find a valid key")
			errors.Add("password", "check password")
			return SerializableValidationErrors{errors}
		}
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

	wrapped := struct {
		Key *model.Key `json:"key"`
	}{key}

	return serialize(wrapped, w)
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
