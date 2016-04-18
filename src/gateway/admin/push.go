package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gateway/config"
	"gateway/core"
	aphttp "gateway/http"
	"gateway/model"
	re "gateway/model/remote_endpoint"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type PushController struct {
	matcher *HostMatcher
	core    *core.Core
}

func RoutePush(controller *PushController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	subscribeRoutes := map[string]http.Handler{
		"PUT": writeForHost(db, controller.Subscribe),
	}
	unsubscribeRoutes := map[string]http.Handler{
		"PUT": writeForHost(db, controller.Unsubscribe),
	}
	publishRoutes := map[string]http.Handler{
		"PUT": writeForHost(db, controller.Publish),
	}
	if conf.CORSEnabled {
		subscribeRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"PUT", "OPTIONS"})
		unsubscribeRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"PUT", "OPTIONS"})
		publishRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"PUT", "OPTIONS"})
	}

	router.Handle(path+"/{endpoint}/subscribe", handlers.MethodHandler(subscribeRoutes)).
		MatcherFunc(controller.matcher.isRouted)
	router.Handle(path+"/{endpoint}/unsubscribe", handlers.MethodHandler(unsubscribeRoutes)).
		MatcherFunc(controller.matcher.isRouted)
	router.Handle(path+"/{endpoint}/publish", handlers.MethodHandler(publishRoutes)).
		MatcherFunc(controller.matcher.isRouted)
}

type Subscription struct {
	Platform string `json:"platform"`
	Channel  string `json:"channel"`
	Period   int64  `json:"period"`
	Name     string `json:"name"`
	Token    string `json:"token"`
}

type Message struct {
	Channel     string                 `json:"channel"`
	Environment string                 `json:"environment"`
	Payload     map[string]interface{} `json:"payload"`
}

func (s *PushController) Subscribe(w http.ResponseWriter, r *http.Request, tx *apsql.Tx, match *HostMatch) aphttp.Error {
	subscription := Subscription{}
	if err := deserialize(&subscription, r.Body); err != nil {
		return err
	}
	if subscription.Platform == "" {
		return aphttp.NewError(errors.New("a platform is required"), http.StatusBadRequest)
	}
	if subscription.Channel == "" {
		return aphttp.NewError(errors.New("a channel is required"), http.StatusBadRequest)
	}
	if subscription.Period <= 0 {
		return aphttp.NewError(errors.New("a period greater than zero is required"), http.StatusBadRequest)
	}
	if subscription.Name == "" {
		return aphttp.NewError(errors.New("a name is required"), http.StatusBadRequest)
	}
	if subscription.Token == "" {
		return aphttp.NewError(errors.New("a token is required"), http.StatusBadRequest)
	}

	codename := mux.Vars(r)["endpoint"]
	endpoint, err := model.FindRemoteEndpointForCodenameAndAPIIDAndAccountID(tx.DB, codename, match.APIID, match.AccountID)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	found := false
	push := &re.Push{}
	err = json.Unmarshal(endpoint.Data, push)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	for _, platform := range push.PushPlatforms {
		if platform.Codename == subscription.Platform {
			found = true
		}
	}
	for _, environment := range endpoint.EnvironmentData {
		push := &re.Push{}
		err = json.Unmarshal(environment.Data, push)
		if err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}
		for _, platform := range push.PushPlatforms {
			if platform.Codename == subscription.Platform {
				found = true
			}
		}
	}
	if !found {
		return aphttp.NewError(fmt.Errorf("%v is not a valid platform", subscription.Platform), http.StatusBadRequest)
	}

	channel := &model.PushChannel{
		AccountID:        match.AccountID,
		APIID:            match.APIID,
		RemoteEndpointID: endpoint.ID,
		Name:             subscription.Channel,
	}
	_channel, err := channel.Find(tx.DB)
	expires := time.Now().Unix() + subscription.Period
	if err != nil {
		channel.Expires = expires
		err := channel.Insert(tx)
		if err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}
	} else {
		channel = _channel
		if channel.Expires < expires {
			channel.Expires = expires
			err := channel.Update(tx)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
		}
	}

	device := &model.PushDevice{
		AccountID:        match.AccountID,
		APIID:            match.APIID,
		RemoteEndpointID: endpoint.ID,
		PushChannelID:    channel.ID,
		Name:             subscription.Name,
	}
	dev, err := device.Find(tx.DB)
	update := false
	if err != nil {
		device.Name = ""
		device.Token = subscription.Token
		dev, err = device.Find(tx.DB)
		if err != nil {
			device.Name = subscription.Name
			device.Type = subscription.Platform
			device.Expires = expires
			err = device.Insert(tx)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
		} else {
			update = true
		}
	} else {
		update = true
	}
	if update {
		dev.Expires = expires
		err := dev.Update(tx)
		if err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}
	}

	return nil
}

func (s *PushController) Unsubscribe(w http.ResponseWriter, r *http.Request, tx *apsql.Tx, match *HostMatch) aphttp.Error {
	subscription := Subscription{}
	if err := deserialize(&subscription, r.Body); err != nil {
		return err
	}
	if subscription.Name == "" && subscription.Token == "" {
		return aphttp.NewError(errors.New("a name or token is required"), http.StatusBadRequest)
	}

	codename := mux.Vars(r)["endpoint"]
	endpoint, err := model.FindRemoteEndpointForCodenameAndAPIIDAndAccountID(tx.DB, codename, match.APIID, match.AccountID)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	channel := &model.PushChannel{
		AccountID:        match.AccountID,
		APIID:            match.APIID,
		RemoteEndpointID: endpoint.ID,
		Name:             subscription.Channel,
	}
	channel, err = channel.Find(tx.DB)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	device := &model.PushDevice{
		AccountID:        match.AccountID,
		APIID:            match.APIID,
		RemoteEndpointID: endpoint.ID,
		PushChannelID:    channel.ID,
		Name:             subscription.Name,
	}
	dev, err := device.Find(tx.DB)
	if err != nil {
		device.Name = ""
		device.Token = subscription.Token
		dev, err = device.Find(tx.DB)
		if err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}
	}
	err = dev.Delete(tx)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	return nil
}

func (s *PushController) Publish(w http.ResponseWriter, r *http.Request, tx *apsql.Tx, match *HostMatch) aphttp.Error {
	message := Message{}
	if err := deserialize(&message, r.Body); err != nil {
		return err
	}
	if message.Channel == "" {
		return aphttp.NewError(errors.New("a channel is required"), http.StatusBadRequest)
	}
	if message.Environment == "" {
		return aphttp.NewError(errors.New("an environment is required"), http.StatusBadRequest)
	}

	codename := mux.Vars(r)["endpoint"]
	endpoint, err := model.FindRemoteEndpointForCodenameAndAPIIDAndAccountID(tx.DB, codename, match.APIID, match.AccountID)
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
		if environment.Name == message.Environment {
			err = json.Unmarshal(environment.Data, epush)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
			break
		}
	}
	epush.UpdateWith(push)
	if !epush.PublishEndpoint {
		return aphttp.NewError(errors.New("publish endpoint is disabled"), http.StatusBadRequest)
	}

	err = s.core.Push.Push(epush, tx, match.AccountID, match.APIID, endpoint.ID, message.Channel, message.Payload)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	return nil
}
