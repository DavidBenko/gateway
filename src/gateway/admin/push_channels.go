package admin

import (
	"encoding/json"
	"net/http"

	"gateway/config"
	"gateway/core"
	aperrors "gateway/errors"
	aphttp "gateway/http"
	"gateway/model"
	re "gateway/model/remote_endpoint"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

type MetaPushChannelsController struct {
	PushChannelsController
	*core.Core
}

func RoutePushChannels(controller ResourceController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	RouteResource(controller, path, router, db, conf)

	routes := map[string]http.Handler{
		"POST": write(db, controller.(*MetaPushChannelsController).Publish),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"POST", "OPTIONS"})
	}

	router.Handle(path+"/{id}/push_manual_messages", handlers.MethodHandler(routes))
}

func (c *MetaPushChannelsController) Publish(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {
	validationErrors := make(aperrors.Errors)
	var wrapped struct {
		Message *Message `json:"push_manual_message"`
	}
	if err := deserialize(&wrapped, r.Body); err != nil {
		return err
	}
	if wrapped.Message.Payload == nil {
		validationErrors.Add("body", "must not be blank")
	} else if len(wrapped.Message.Payload) == 0 {
		validationErrors.Add("body", "must contain one playload")
	}
	if wrapped.Message.Environment == "" {
		validationErrors.Add("environment", "must not be blank")
	}
	if !validationErrors.Empty() {
		return SerializableValidationErrors{validationErrors}
	}

	object := model.PushChannel{}
	c.mapFields(r, &object)
	channel, err := object.Find(tx.DB)
	if err != nil {
		return c.notFound()
	}

	endpoint, err := model.FindRemoteEndpointForAPIIDAndAccountID(tx.DB, channel.RemoteEndpointID, channel.APIID, channel.AccountID)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	push := &re.Push{}
	err = json.Unmarshal(endpoint.Data, push)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	epush := &re.Push{}
	for _, environment := range endpoint.EnvironmentData {
		if environment.Name == wrapped.Message.Environment {
			err = json.Unmarshal(environment.Data, epush)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
			break
		}
	}
	epush.UpdateWith(push)

	err = c.Push.Push(epush, tx, channel.AccountID, channel.APIID, endpoint.ID, channel.Name, wrapped.Message.Payload)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	return nil
}
