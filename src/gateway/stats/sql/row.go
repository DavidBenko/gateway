package sql

import "gateway/stats"

// Row is a wrapper for stats.Row in sql.
type Row struct {
	stats.Row
	APIID                      int64  `db:"api_id"`
	APIName                    string `db:"api_name"`
	HostID                     int64  `db:"host_id"`
	HostName                   string `db:"host_name"`
	ProxyEnvID                 int64  `db:"proxy_env_id"`
	ProxyEnvName               string `db:"proxy_env_name"`
	ProxyGroupID               int64  `db:"proxy_group_id"`
	ProxyGroupName             string `db:"proxy_group_name"`
	ProxyID                    int64  `db:"proxy_id"`
	ProxyName                  string `db:"proxy_name"`
	ProxyRoutePath             string `db:"proxy_route_path"`
	ProxyRouteVerb             string `db:"proxy_route_verb"`
	RemoteEndpointResponseTime int    `db:"remote_endpoint_response_time"`
	RequestID                  string `db:"request_id"`
	RequestSize                int    `db:"request_size"`
	ResponseError              string `db:"response_error"`
	ResponseSize               int    `db:"response_size"`
	ResponseStatus             int    `db:"response_status"`
	ResponseTime               int    `db:"response_time"`
}

func (r *Row) value(k string) interface{} {
	return map[string]interface{}{
		"api.id":                        r.APIID,
		"api.name":                      r.APIName,
		"host.id":                       r.HostID,
		"host.name":                     r.HostName,
		"proxy.env.id":                  r.ProxyEnvID,
		"proxy.env.name":                r.ProxyEnvName,
		"proxy.group.id":                r.ProxyGroupID,
		"proxy.group.name":              r.ProxyGroupName,
		"proxy.id":                      r.ProxyID,
		"proxy.name":                    r.ProxyName,
		"proxy.route.path":              r.ProxyRoutePath,
		"proxy.route.verb":              r.ProxyRouteVerb,
		"remote_endpoint.response.time": r.RemoteEndpointResponseTime,
		"request.size":                  r.RequestSize,
		"request.id":                    r.RequestID,
		"response.time":                 r.ResponseTime,
		"response.size":                 r.ResponseSize,
		"response.status":               r.ResponseStatus,
		"response.error":                r.ResponseError,
	}[k]
}
