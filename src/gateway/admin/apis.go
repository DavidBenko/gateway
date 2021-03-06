package admin

import (
	"errors"
	"fmt"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	"gateway/names"
	apsql "gateway/sql"
	"net/http"
	"strings"

	"github.com/aymerick/raymond"
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
		logreport.Printf("%s Error importing api: %v", config.System, err)
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
		logreport.Printf("Unable to decode export due to error: %v", err)
		return errors.New("Unable to decode export.")
	}

	newAPI.Name = api.Name
	newAPI.ID = api.ID
	newAPI.AccountID = api.AccountID
	newAPI.UserID = api.UserID

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

	var host *model.Host
	var err error
	if host, err = c.addDefaultHost(api, tx); err != nil {
		return err
	}

	if host != nil {
		api.Hosts = []*model.Host{host}
	}

	return nil
}

func (c *APIsController) addDefaultEnvironment(api *model.API, tx *apsql.Tx) error {
	if c.conf.AddDefaultEnvironment {
		env := &model.Environment{Name: c.conf.DefaultEnvironmentName}
		env.AccountID = api.AccountID
		env.UserID = api.UserID
		env.APIID = api.ID
		env.SessionType = model.SessionTypeClient
		env.SessionHeader = model.SessionHeaderDefault

		if err := env.Insert(tx); err != nil {
			return err
		}
	}

	return nil
}

func (c *APIsController) addDefaultHost(api *model.API, tx *apsql.Tx) (*model.Host, error) {
	if !c.conf.CreateDefaultHost {
		return nil, nil
	}

	generatedHostName := names.GenerateHostName()

	host := &model.Host{Name: generatedHostName, Hostname: fmt.Sprintf("%s.%s", generatedHostName, defaultDomain)}
	host.AccountID = api.AccountID
	host.UserID = api.UserID
	host.APIID = api.ID

	if err := host.Insert(tx); err != nil {
		return nil, err
	}

	return host, nil
}

// AfterUpdate does some work after updating a record in the database
func (c *APIsController) AfterUpdate(api *model.API, tx *apsql.Tx) error {
	return c.populateHosts(api, tx.DB)
}

// AfterFind does some work after finding a record in the database
func (c *APIsController) AfterFind(api *model.API, db *apsql.DB) error {
	return c.populateHosts(api, db)
}

func (c *APIsController) populateHosts(api *model.API, db *apsql.DB) error {
	hosts, err := model.AllHostsForAPIIDAndAccountID(db, api.ID, api.AccountID)
	if err != nil {
		return err
	}

	api.Hosts = hosts

	return nil
}

type enhancedAPI struct {
	*model.API
	BaseURL string `json:"base_url"`
}

func (c *APIsController) addBaseURL(api *model.API) *enhancedAPI {
	hosts := make([]string, len(api.Hosts), len(api.Hosts))
	for idx, host := range api.Hosts {
		hosts[idx] = host.Hostname
	}
	interpolated, err := c.interpolateDefaultAPIAccessScheme(hosts)
	if err != nil {
		logreport.Printf("Encountered error attempting to interpolate hosts to produce API base URL: %v", err)
	}
	return &enhancedAPI{API: api, BaseURL: interpolated}
}

func (c *APIsController) interpolateDefaultAPIAccessScheme(hosts []string) (string, error) {
	if hosts == nil || len(hosts) == 0 {
		return "", nil
	}

	if !strings.Contains(c.conf.DefaultAPIAccessScheme, "{{") && !strings.Contains(c.conf.DefaultAPIAccessScheme, "}}") {
		return c.conf.DefaultAPIAccessScheme, nil
	}

	return interpolate(c.conf.DefaultAPIAccessScheme, map[string]interface{}{"hosts": hosts})
}

func interpolate(expression string, ctx map[string]interface{}) (string, error) {
	templ, err := raymond.Parse(expression)
	if err != nil {
		return "", err
	}

	result, err := templ.Exec(ctx)
	if err != nil {
		return "", err
	}

	return result, nil
}
