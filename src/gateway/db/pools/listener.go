package pools

import (
	"log"

	"gateway/db"
	"gateway/sql"
)

// Notify implements gateway/sql Listener Notify to flush db entries.
func (p *Pools) Notify(notif *sql.Notification) {
	if notif.Event != sql.Delete || notif.Table != "remote_endpoints" {
		return
	}

	for _, msg := range notif.Messages {
		p.flushByMsg(msg)
	}
}

// flushByMsg flushes the entry for a db.Specifier contained in a Message.
func (p *Pools) flushByMsg(msg interface{}) {
	switch m := msg.(type) {
	case db.Specifier:
		pool, err := p.poolForSpec(m)
		if err != nil {
			log.Printf("error flushing DB cache: %s", err.Error())
		}
		FlushEntry(pool, m)
	case error:
		log.Printf("tried to flush db entry but received error: %s", m.Error())
	default:
	}
}

// Reconnect implements gateway/sql Listener Reconnect by flushing all db's.
func (p *Pools) Reconnect() {
	for _, pool := range []*serverPool{
		p.sqlPool,
	} {
		for _, db := range pool.dbs {
			FlushEntry(pool, db.Spec())
		}
	}
}
