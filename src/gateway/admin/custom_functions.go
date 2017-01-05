package admin

import (
	"archive/tar"
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

	input, output := &bytes.Buffer{}, &bytes.Buffer{}
	image := tar.NewWriter(input)

	for _, file := range files {
		header := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := image.WriteHeader(header); err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}
		if _, err := image.Write([]byte(file.Body)); err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}
	}

	if err := image.Close(); err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	options := dockerclient.BuildImageOptions{
		Name:         function.ImageName(),
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
