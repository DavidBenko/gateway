package admin

import (
	"encoding/json"
	"time"

	"gateway/config"
	"gateway/docker"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	dockerclient "github.com/fsouza/go-dockerclient"
	"golang.org/x/net/websocket"
)

func (c *CustomFunctionsController) AfterInsert(function *model.CustomFunction, tx *apsql.Tx) error {
	return function.AfterInsert(tx)
}

type CustomFunctionBuildController struct {
	BaseController
	db *apsql.DB
}

func RouteCustomFunctionBuild(controller *CustomFunctionBuildController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {
	controller.db = db
	router.Handle(path, websocket.Handler(controller.Build))
}

type CustomFunctionBuildResult struct {
	Time int64 `json:"time"`
}

type CustomFunctionLogLine struct {
	Line string `json:"line"`
}

type DockerWriter struct {
	ws *websocket.Conn
}

func (w *DockerWriter) Write(p []byte) (n int, err error) {
	line := &CustomFunctionLogLine{
		Line: string(p),
	}

	wrapped := struct {
		Line *CustomFunctionLogLine `json:"line"`
	}{line}

	body, err := json.Marshal(&wrapped)
	if err != nil {
		return 0, err
	}

	return w.ws.Write(body)
}

func (c *CustomFunctionBuildController) Build(ws *websocket.Conn) {
	db, r := c.db, ws.Request()
	accountID, apiID, customFunctionID := c.accountID(r), apiIDFromPath(r), customFunctionIDFromPath(r)

	customFunction := model.CustomFunction{
		AccountID: accountID,
		APIID:     apiID,
		ID:        customFunctionID,
	}
	function, err := customFunction.Find(db)
	if err != nil {
		ws.Close()
		return
	}

	file := model.CustomFunctionFile{
		AccountID:        function.AccountID,
		APIID:            function.APIID,
		CustomFunctionID: function.ID,
	}
	files, err := file.All(db)
	if err != nil {
		ws.Close()
		return
	}

	input, err := files.Tar()
	if err != nil {
		ws.Close()
		return
	}

	output := &DockerWriter{
		ws: ws,
	}
	options := dockerclient.BuildImageOptions{
		Name:         function.ImageName(),
		NoCache:      true,
		InputStream:  input,
		OutputStream: output,
	}

	start := time.Now()
	err = docker.BuildImage(options)
	if err != nil {
		ws.Close()
		return
	}

	result := &CustomFunctionBuildResult{
		Time: (time.Since(start).Nanoseconds() + +5e5) / 1e6,
	}

	wrapped := struct {
		Result *CustomFunctionBuildResult `json:"result"`
	}{result}

	body, err := json.Marshal(&wrapped)
	if err != nil {
		ws.Close()
		return
	}

	_, err = ws.Write(body)
	if err != nil {
		ws.Close()
		return
	}

	ws.Close()

	return
}
