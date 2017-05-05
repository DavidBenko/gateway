package admin

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

func RouteCustomFunctionTest(controller *CustomFunctionTestController, path string,
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

type CustomFunctionTestController struct {
	BaseController
}

type CustomFunctionTestResult struct {
	Output     string `json:"output"`
	Log        string `json:"log"`
	Time       int64  `json:"time"`
	StatusCode int64  `json:"status_code"`
	Error      string `json:"error,omitempty"`
}

func (c *CustomFunctionTestController) Test(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
	accountID, apiID, customFunctionID, testID := c.accountID(r), apiIDFromPath(r), customFunctionIDFromPath(r), testIDFromPath(r)

	customFunctionTest := model.CustomFunctionTest{
		AccountID:        accountID,
		APIID:            apiID,
		CustomFunctionID: customFunctionID,
		ID:               testID,
	}
	test, err := customFunctionTest.Find(db)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	var input interface{}
	err = json.Unmarshal(test.Input, &input)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	start := time.Now()
	runOutput, err := model.ExecuteCustomFunction(db, accountID, apiID, customFunctionID, "", input, false)
	var customFunctionErr string
	if err != nil {
		customFunctionErr = err.Error()
	}
	elapsed := (time.Since(start).Nanoseconds() + +5e5) / 1e6
	lines, output := runOutput.Parts()
	result := struct {
		Result CustomFunctionTestResult `json:"result"`
	}{
		CustomFunctionTestResult{
			Output:     output,
			Log:        strings.Join(lines, "\n"),
			Time:       elapsed,
			StatusCode: int64(runOutput.StatusCode),
			Error:      customFunctionErr,
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
