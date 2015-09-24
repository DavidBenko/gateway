package admin

import (
	"encoding/json"

	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	"golang.org/x/net/websocket"
)

const (
	NOTIFY_COMMAND_REGISTER = iota
	NOTIFY_COMMAND_UNREGISTER
)

type Notification struct {
	Resource   string `json:"resource"`
	Action     string `json:"action"`
	ResourceID int64  `json:"resource_id"`
	APIID      int64  `json:"api_id"`
	User       string `json:"user"`
}

var RESOURCE_MAP = map[string]string{
	"accounts":         "account",
	"apis":             "api",
	"endpoint_groups":  "endpoint-group",
	"environments":     "environment",
	"hosts":            "host",
	"libraries":        "library",
	"proxy_endpoints":  "proxy-endpoint",
	"remote_endpoints": "remote-endpoint",
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

func RouteNotify(notify *NotifyController, path string, router aphttp.Router, db *apsql.DB) {
	notify.notifications = make(chan *apsql.Notification, 8)
	notify.command = make(chan *NotifyCommand, 8)
	notify.db = db
	router.Handle(path, websocket.Handler(notify.NotifyHandler))
	db.RegisterListener(notify)
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
		n := &Notification{
			Resource:   RESOURCE_MAP[notification.Table],
			Action:     ACTION_MAP[notification.Event],
			ResourceID: int64(notification.ID),
			APIID:      int64(notification.APIID),
			User:       email,
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
