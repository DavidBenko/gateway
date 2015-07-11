package sql

import "gateway/stats"

// Row is a wrapper for stats.Row in sql.
type Row struct {
	stats.Row
	RequestSize    int    `db:"request_size"`
	RequestID      string `db:"request_id"`
	ResponseTime   int    `db:"response_time"`
	ResponseSize   int    `db:"response_size"`
	ResponseStatus int    `db:"response_status"`
	ResponseError  string `db:"response_error"`
}

func (r *Row) value(k string) interface{} {
	return map[string]interface{}{
		"request.size":    r.RequestSize,
		"request.id":      r.RequestID,
		"response.time":   r.ResponseTime,
		"response.size":   r.ResponseSize,
		"response.status": r.ResponseStatus,
		"response.error":  r.ResponseError,
	}[k]
}
