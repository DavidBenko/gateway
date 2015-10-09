package admin

import (
	"errors"
	"fmt"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/vincent-petithory/dataurl"
)

// RouteAPIExport routes the endpoint for API export
func RouteAPIExport(controller *APIsController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET": read(db, controller.Export),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

// Export exports a whole API
func (c *APIsController) Export(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {

	id := instanceID(r)

	api, err := model.FindAPIForAccountIDForExport(db, id, c.accountID(r))

	if err != nil {
		return c.notFound()
	}

	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s.json"`, api.Name))

	return c.serializeInstance(api, w)
}

func (c *APIsController) decodeExport(api *model.API, tx *apsql.Tx) (*model.API, aphttp.Error) {
	decoded, err := dataurl.DecodeString(api.Export)
	if err != nil {
		return nil, aphttp.NewError(fmt.Errorf("Encountered an error attempting to decode export: %v", err), http.StatusBadRequest)
	}

	api, httpErr := c.deserializeInstance(strings.NewReader(string(decoded.Data)))
	if httpErr != nil {
		newErr := aphttp.NewError(fmt.Errorf("API file is invalid: %s", httpErr), httpErr.Code())
		return nil, newErr
	}

	return api, httpErr
}

// Import imports a full API
func (c *APIsController) importAPI(newAPI *model.API, tx *apsql.Tx) aphttp.Error {
	if newAPI.Export == "" {
		return nil
	}

	api, err := c.decodeExport(newAPI, tx)
	if err != nil {
		log.Printf("Unable to decode export due to error: %v", err)
		return aphttp.NewError(errors.New("Unable to decode export."),
			http.StatusBadRequest)
	}

	api.Name = newAPI.Name
	api.ID = newAPI.ID
	api.AccountID = newAPI.AccountID

	validationErrors := api.Validate()
	if !validationErrors.Empty() {
		return SerializableValidationErrors{validationErrors}
	}

	if err := api.Import(tx); err != nil {
		validationErrors = api.ValidateFromDatabaseError(err)
		if !validationErrors.Empty() {
			return SerializableValidationErrors{validationErrors}
		}
		log.Printf("%s Error importing api: %v", config.System, err)
		return aphttp.NewServerError(err)
	}

	newAPI.CopyFrom(api, false)

	if err := c.addLocalhost(api, tx); err != nil {
		return aphttp.DefaultServerError()
	}

	return nil
}

// AfterInsert does some work after inserting an API
func (c *APIsController) AfterInsert(api *model.API, tx *apsql.Tx) error {
	if api.Export == "" {
		if err := c.addDefaultEnvironment(api, tx); err != nil {
			return err
		}
		if err := c.addLocalhost(api, tx); err != nil {
			return err
		}
	} else {
		if err := c.importAPI(api, tx); err != nil {
			return err.Error()
		}
	}

	return nil
}

func (c *APIsController) addDefaultEnvironment(api *model.API, tx *apsql.Tx) error {
	if !c.conf.DevMode {
		return nil
	}

	if c.conf.AddDefaultEnvironment {
		env := &model.Environment{Name: c.conf.DefaultEnvironmentName}
		env.AccountID = api.AccountID
		env.APIID = api.ID

		if err := env.Insert(tx); err != nil {
			return err
		}
	}

	return nil
}

func (c *APIsController) addLocalhost(api *model.API, tx *apsql.Tx) error {
	if !c.conf.DevMode {
		return nil
	}

	if c.conf.AddLocalhost {
		any, err := model.AnyHostExists(tx)
		if err != nil {
			return err
		}
		if !any {
			host := &model.Host{Name: "localhost", Hostname: "localhost"}
			host.AccountID = api.AccountID
			host.APIID = api.ID

			if err := host.Insert(tx); err != nil {
				return err
			}
		}
	}

	return nil
}
