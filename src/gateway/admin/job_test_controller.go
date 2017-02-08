package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gateway/config"
	"gateway/core"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

func RouteJobTest(controller *JobTestController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET":  read(db, controller.Test),
		"POST": read(db, controller.Test),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "POST", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

type JobTestController struct {
	BaseController
	*core.Core
}

type JobTestResult struct {
	Log   string `json:"log"`
	Time  int64  `json:"time"`
	Error string `json:"error,omitempty"`
}

type JobTestResponse struct {
	Result JobTestResult `json:"result"`
}

func (c *JobTestController) Test(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
	accountID, apiID, endpointID, testID := c.accountID(r), apiIDFromPath(r), endpointIDFromPath(r), testIDFromPath(r)

	jobTest := model.JobTest{
		AccountID: accountID,
		APIID:     apiID,
		JobID:     endpointID,
		ID:        testID,
	}
	test, err := jobTest.Find(db)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	proxyEndpoint := model.ProxyEndpoint{
		AccountID: accountID,
		APIID:     apiID,
		ID:        endpointID,
		Type:      model.ProxyEndpointTypeJob,
	}
	endpoint, err := proxyEndpoint.Find(db)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	logs := &bytes.Buffer{}
	logPrint := logreport.PrintfCopier(logs)

	logPrefix := fmt.Sprintf("%s [act %d] [api %d] [end %d]", config.Job,
		endpoint.AccountID, endpoint.APIID, endpoint.ID)

	parametersJSON, err := json.Marshal(test.Parameters)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	start := time.Now()
	err = c.ExecuteJob(endpoint.ID, endpoint.AccountID, endpoint.APIID, logPrint, logPrefix, string(parametersJSON))
	var jobError string
	if err != nil {
		jobError = err.Error()
	}
	elapsed := (time.Since(start).Nanoseconds() + +5e5) / 1e6
	result := JobTestResponse{
		JobTestResult{
			Log:   logs.String(),
			Time:  elapsed,
			Error: jobError,
		},
	}

	body, err := json.MarshalIndent(&result, "", "    ")
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

	return nil
}
