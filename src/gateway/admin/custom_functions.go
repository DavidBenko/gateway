package admin

import (
	"bytes"
	"encoding/json"
	"net/http"

	"gateway/config"
	"gateway/docker"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/handlers"
)

func (c *CustomFunctionsController) AfterInsert(function *model.CustomFunction, tx *apsql.Tx) error {
	return function.AfterInsert(tx)
}

type CustomFunctionBuildController struct {
	BaseController
}

func RouteCustomFunctionBuild(controller *CustomFunctionBuildController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET": read(db, controller.Build),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

type CustomFunctionBuildResult struct {
	Log  string `json:"log"`
	Time int64  `json:"time"`
}

func (c *CustomFunctionBuildController) Build(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
	accountID, apiID, customFunctionID := c.accountID(r), apiIDFromPath(r), customFunctionIDFromPath(r)

	customFunction := model.CustomFunction{
		AccountID: accountID,
		APIID:     apiID,
		ID:        customFunctionID,
	}
	function, err := customFunction.Find(db)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	file := model.CustomFunctionFile{
		AccountID:        function.AccountID,
		APIID:            function.APIID,
		CustomFunctionID: function.ID,
	}
	files, err := file.All(db)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	input, err := files.Tar()
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	output := &bytes.Buffer{}
	options := dockerclient.BuildImageOptions{
		Name:         function.ImageName(),
		NoCache:      true,
		InputStream:  input,
		OutputStream: output,
	}

	err = docker.BuildImage(options)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	result := &CustomFunctionBuildResult{
		Log:  output.String(),
		Time: 0,
	}

	wrapped := struct {
		Result *CustomFunctionBuildResult `json:"result"`
	}{result}

	body, err := json.MarshalIndent(&wrapped, "", "    ")
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

	return nil
}
