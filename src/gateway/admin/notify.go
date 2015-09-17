package admin

import (
	"encoding/json"
	"strconv"

	aphttp "gateway/http"
	apsql "gateway/sql"

	"golang.org/x/net/websocket"
)

const (
	NOTIFY_COMMAND_REGISTER = iota
	NOTIFY_COMMAND_UNREGISTER
)

type NotifyCommand struct {
	Command   int
	AccountID int
	Comm      chan *apsql.Notification
}

type Notify struct {
	notifications chan *apsql.Notification
	command       chan *NotifyCommand
}

func RouteNotify(path string, router aphttp.Router, db *apsql.DB) {
	notify := &Notify{
		notifications: make(chan *apsql.Notification, 8),
		command:       make(chan *NotifyCommand, 8),
	}
	router.Handle(path, websocket.Handler(notify.NotifyHandler))
	db.RegisterListener(notify)
	go notify.Queue()
}

func (n *Notify) Notify(notification *apsql.Notification) {
	n.notifications <- notification
}

func (n *Notify) Reconnect() {

}

func (n *Notify) Queue() {
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

func (n *Notify) NotifyHandler(ws *websocket.Conn) {
	request := ws.Request()
	request.ParseForm()
	if len(request.Form["account_id"]) != 1 {
		return
	}
	id, err := strconv.Atoi(request.Form["account_id"][0])
	if err != nil {
		return
	}
	register := &NotifyCommand{
		Command:   NOTIFY_COMMAND_REGISTER,
		AccountID: id,
		Comm:      make(chan *apsql.Notification, 8),
	}
	n.command <- register
	defer func() {
		unregister := &NotifyCommand{
			Command:   NOTIFY_COMMAND_UNREGISTER,
			AccountID: id,
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
		json, err := json.Marshal(notification)
		if err != nil {
			return
		}
		_, err = ws.Write(json)
		if err != nil {
			return
		}
	}
}
