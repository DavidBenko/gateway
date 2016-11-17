package admin

import (
	"encoding/json"

	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"
	"gateway/store"

	"golang.org/x/net/websocket"
)

const (
	NOTIFY_COMMAND_REGISTER = iota
	NOTIFY_COMMAND_UNREGISTER
)

type Notification struct {
	Resource        string `json:"resource"`
	Action          string `json:"action"`
	ResourceID      int64  `json:"resource_id"`
	ProxyEndpointID int64  `json:"proxy_endpoint_id"`
	APIID           int64  `json:"api_id"`
	User            string `json:"user"`
	Tag             string `json:"tag"`
}

var RESOURCE_MAP = map[string]string{
	"accounts":                  "account",
	"apis":                      "api",
	"collections":               "store-collection",
	"endpoint_groups":           "endpoint-group",
	"environments":              "environment",
	"hosts":                     "host",
	"libraries":                 "library",
	"objects":                   "store-object",
	"proxy_endpoints":           "proxy-endpoint",
	"proxy_endpoint_schemas":    "proxy-endpoint-schema",
	"remote_endpoints":          "remote-endpoint",
	"scratch_pads":              "remote-endpoint-environment-datum-scratch-pad",
	"users":                     "user",
	"proxy_endpoint_components": "shared-component",
	"push_channels":             "push-channel",
	"push_channels_devices":     "push-channel-device",
	"push_channel_messages":     "push-channel-message",
	"push_messages":             "push-channel-device-message",
	"push_devices":              "push-device",
	"jobs":                      "job",
	"timers":                    "timer",
	"job_tests":                 "job-test",
}

var ACTION_MAP = [...]string{
	"create",
	"update",
	"delete",
}

type NotifyCommand struct {
	Command   int
	AccountID int
	Comm      chan *apsql.Notification
}

type NotifyController struct {
	BaseController
	notifications chan *apsql.Notification
	command       chan *NotifyCommand
	db            *apsql.DB
}

func RouteNotify(notify *NotifyController, path string, router aphttp.Router, db *apsql.DB, s store.Store) {
	notify.notifications = make(chan *apsql.Notification, 1024)
	notify.command = make(chan *NotifyCommand, 8)
	notify.db = db
	router.Handle(path, websocket.Handler(notify.NotifyHandler))
	db.RegisterListener(notify)
	s.RegisterListener(notify)
	go notify.Queue()
}

func (n *NotifyController) Notify(notification *apsql.Notification) {
	n.notifications <- notification
}

func (n *NotifyController) Reconnect() {

}

func (n *NotifyController) Queue() {
	clients := make([]*NotifyCommand, 8)
	for {
		select {
		case notification := <-n.notifications:
			for _, client := range clients {
				if client == nil {
					continue
				}
				if client.AccountID == int(notification.AccountID) {
					client.Comm <- notification
				}
			}
		case command := <-n.command:
			switch command.Command {
			case NOTIFY_COMMAND_REGISTER:
				found := false
				for k, v := range clients {
					if v == nil {
						clients[k], found = command, true
						break
					}
				}
				if !found {
					clients = append(clients, command)
				}
			case NOTIFY_COMMAND_UNREGISTER:
				for k, v := range clients {
					if v == nil {
						continue
					}
					if v.Comm == command.Comm {
						clients[k] = nil
						close(command.Comm)
						break
					}
				}
			}
		}
	}
}

func (n *NotifyController) NotifyHandler(ws *websocket.Conn) {
	request := ws.Request()
	account := int(n.accountID(request))
	register := &NotifyCommand{
		Command:   NOTIFY_COMMAND_REGISTER,
		AccountID: account,
		Comm:      make(chan *apsql.Notification, 8),
	}
	n.command <- register
	defer func() {
		unregister := &NotifyCommand{
			Command:   NOTIFY_COMMAND_UNREGISTER,
			AccountID: account,
			Comm:      register.Comm,
		}
		go func() {
			for _ = range register.Comm {
				//noop
			}
		}()
		n.command <- unregister
	}()
	for notification := range register.Comm {
		email := "unknown"
		user, err := model.FindUserByID(n.db, notification.UserID)
		if err == nil {
			email = user.Email
		}

		resource, found := RESOURCE_MAP[notification.Table]
		// only send notifications that we're interested in, that have a corresponding
		// resource found in the resource map
		if found {
			n := &Notification{
				Resource:        resource,
				Action:          ACTION_MAP[notification.Event],
				ResourceID:      int64(notification.ID),
				ProxyEndpointID: int64(notification.ProxyEndpointID),
				APIID:           int64(notification.APIID),
				User:            email,
				Tag:             notification.Tag,
			}

			json, err := json.Marshal(n)
			if err != nil {
				return
			}

			_, err = ws.Write(json)
			if err != nil {
				return
			}
		}
	}
}
