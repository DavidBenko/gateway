package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"gateway/config"
	"gateway/core"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

type MetaScratchPadsController struct {
	ScratchPadsController
	*core.Core
}

func RouteScratchPads(controller ResourceController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	RouteResource(controller, path, router, db, conf)

	routes := map[string]http.Handler{
		"GET": read(db, controller.(*MetaScratchPadsController).Test),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path+"/{id}/test", handlers.MethodHandler(routes))
}

type ScratchPadResult struct {
	Request  string `json:"request"`
	Response string `json:"response"`
	Time     int64  `json:"time"`
}

func (c *MetaScratchPadsController) Test(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
	object := model.ScratchPad{}
	c.mapFields(r, &object)
	pad, err := object.Find(db)
	if err != nil {
		return c.notFound()
	}

	endpoint, err := model.FindRemoteEndpointForAPIIDAndAccountID(db, pad.RemoteEndpointID, pad.APIID, pad.AccountID)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	for _, data := range endpoint.EnvironmentData {
		if data.ID == pad.RemoteEndpointEnvironmentDataID {
			endpoint.SelectedEnvironmentData = &data.Data
			break
		}
	}

	vm := core.VMCopy()
	_, err = vm.Run(pad.Code)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	result := ScratchPadResult{}
	obj, err := vm.Run("JSON.stringify(request);")
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	var rawRequest json.RawMessage
	err = json.Unmarshal([]byte(obj.String()), &rawRequest)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	request, err := c.PrepareRequest(endpoint, &rawRequest)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	jsonRequest, err := request.JSON()
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	result.Request = string(jsonRequest)
	start := time.Now()
	response := request.Perform()
	result.Time = (time.Since(start).Nanoseconds() + 5e5) / 1e6
	jsonResponse, err := response.JSON()
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	result.Response = string(jsonResponse)

	body, err := json.MarshalIndent(&result, "", "    ")
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

	return nil
}
