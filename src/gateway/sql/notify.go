package sql

import (
	"encoding/json"
	"fmt"
	"gateway/config"
	"log"
	"time"

	"github.com/lib/pq"
)

type ChangeEventType int

const (
	Insert ChangeEventType = iota
	Update
	Delete
)

const postgresNotifyChannel = "gateway"

type Notification struct {
	Table string
	APIID int64
	Event ChangeEventType
}

type Listener interface {
	Notify(*Notification)
	Reconnect()
}

func (db *DB) RegisterListener(l Listener) {
	defer db.listenersMutex.Unlock()
	db.listenersMutex.Lock()
	db.listeners = append(db.listeners, l)
}

func (db *DB) NotifyListeners(n *Notification) {
	defer db.listenersMutex.RUnlock()
	db.listenersMutex.RLock()

	for _, listener := range db.listeners {
		listener.Notify(n)
	}
}

func (db *DB) NotifyListenersOfReconnection() {
	defer db.listenersMutex.RUnlock()
	db.listenersMutex.RLock()

	for _, listener := range db.listeners {
		listener.Reconnect()
	}
}

func (tx *Tx) Notify(table string, apiID int64, event ChangeEventType) error {
	n := Notification{table, apiID, event}

	// Notifications should be handled manually
	// after transaction committal in SQLite3.
	// This is for development only, on a single box.
	if tx.db.Driver == Sqlite3 {
		tx.queueNotification(&n)
		return nil
	}

	// However, they should be sent to the DB
	// within a transaction for Postgres, so that
	// the fire on COMMIT for all listeners.
	if tx.db.Driver == Postgres {
		json, err := json.Marshal(&n)
		if err != nil {
			return err
		}
		_, err = tx.Exec(fmt.Sprintf("Notify \"%s\", '%s'",
			postgresNotifyChannel, string(json)))
		return err
	}

	return nil
}

func (tx *Tx) queueNotification(n *Notification) {
	tx.notifications = append(tx.notifications, n)
}

// Commit commits the transaction and sends out pending notifications
func (tx *Tx) Commit() error {
	err := tx.Tx.Commit()
	if err == nil {
		for _, n := range tx.notifications {
			tx.db.NotifyListeners(n)
		}
	}
	return err
}

func (db *DB) startListening(conf config.Database) error {
	listener := pq.NewListener(conf.ConnectionString,
		2*time.Second,
		time.Minute,
		db.listenerConnectionEvent)
	err := listener.Listen(postgresNotifyChannel)
	if err != nil {
		return err
	}
	go db.waitForNotification(listener)
	return nil
}

func (db *DB) waitForNotification(l *pq.Listener) {
	for {
		select {
		case pgNotification := <-l.Notify:
			if pgNotification.Channel == postgresNotifyChannel {
				var notification Notification
				err := json.Unmarshal([]byte(pgNotification.Extra), &notification)
				if err != nil {
					log.Printf("%s Error parsing notification '%s': %v",
						config.System, pgNotification.Extra, err)
					continue
				}
				db.NotifyListeners(&notification)
			} else {
				db.NotifyListenersOfReconnection()
			}
		case <-time.After(90 * time.Second):
			go func() {
				l.Ping()
			}()
		}
	}
}

func (db *DB) listenerConnectionEvent(ev pq.ListenerEventType, err error) {
	if err != nil {
		log.Printf("%s Database listener connection problem: %v", config.System, err)
	}
}
