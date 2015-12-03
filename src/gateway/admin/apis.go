package admin

import (
	"errors"
	"fmt"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/logger"
	"gateway/model"
	"gateway/names"
	apsql "gateway/sql"
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
func (c *APIsController) importAPI(api *model.API, tx *apsql.Tx) aphttp.Error {
	if err := api.Import(tx); err != nil {
		validationErrors := api.ValidateFromDatabaseError(err)
		if !validationErrors.Empty() {
			return SerializableValidationErrors{validationErrors}
		}
		logger.Printf("%s Error importing api: %v", config.System, err)
		return aphttp.NewServerError(err)
	}

	return nil
}

func (c *APIsController) BeforeValidate(api *model.API, tx *apsql.Tx) error {
	if api.Export == "" {
		return nil
	}

	newAPI, err := c.decodeExport(api, tx)
	if err != nil {
		logger.Printf("Unable to decode export due to error: %v", err)
		return errors.New("Unable to decode export.")
	}

	newAPI.Name = api.Name
	newAPI.ID = api.ID
	newAPI.AccountID = api.AccountID

	*api = *newAPI
	return nil
}

// AfterInsert does some work after inserting an API
func (c *APIsController) AfterInsert(api *model.API, tx *apsql.Tx) error {
	fromExport := api.ExportVersion > 0
	if !fromExport {
		tx.PushTag(apsql.NotificationTagAuto)
		defer tx.PopTag()
		if err := c.addDefaultEnvironment(api, tx); err != nil {
			return err
		}
	} else {
		tx.PushTag(apsql.NotificationTagImport)
		defer tx.PopTag()
		if err := c.importAPI(api, tx); err != nil {
			return err.Error()
		}
		api.Normalize()
	}

	if err := c.addDefaultHost(api, tx); err != nil {
		return err
	}

	return nil
}

func (c *APIsController) addDefaultEnvironment(api *model.API, tx *apsql.Tx) error {
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

func (c *APIsController) addDefaultHost(api *model.API, tx *apsql.Tx) error {
	if !c.conf.CreateDefaultHost {
		return nil
	}

	generatedHostName := names.GenerateHostName()

	host := &model.Host{Name: generatedHostName, Hostname: fmt.Sprintf("%s.%s", generatedHostName, defaultDomain)}
	host.AccountID = api.AccountID
	host.APIID = api.ID

	if err := host.Insert(tx); err != nil {
		return err
	}

	return nil
}
