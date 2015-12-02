package pools

import (
	logger "log"

	"gateway/db"
	"gateway/sql"
)

// Notify implements gateway/sql Listener Notify to flush db entries.
func (p *Pools) Notify(notif *sql.Notification) {
	if notif.Table != "remote_endpoints" {
		return
	}

	switch notif.Event {
	case sql.Delete, sql.Update:
		// Pass through to flush
	default:
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
			logger.Printf("error flushing DB cache: %s", err.Error())
		}
		pool.Lock()
		FlushEntry(pool, m)
		pool.Unlock()
	case error:
		logger.Printf("tried to flush db entry but received error: %s", m.Error())
	default:
	}
}

// Reconnect implements gateway/sql Listener Reconnect by flushing all db's.
func (p *Pools) Reconnect() {
	for _, pool := range []ServerPool{
		p.sqlsPool,
		p.pqPool,
		p.mongoPool,
	} {
		pool.Lock()
		for spec := range pool.Iterator() {
			FlushEntry(pool, spec)
		}
		pool.Unlock()
	}
}
