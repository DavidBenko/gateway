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

// value gets the value of r mapped by its stats variable name.
func (r *Row) value(k string) interface{} {
	switch k {
	case "api.id":
		return r.APIID
	case "api.name":
		return r.APIName
	case "host.id":
		return r.HostID
	case "host.name":
		return r.HostName
	case "proxy.env.id":
		return r.ProxyEnvID
	case "proxy.env.name":
		return r.ProxyEnvName
	case "proxy.group.id":
		return r.ProxyGroupID
	case "proxy.group.name":
		return r.ProxyGroupName
	case "proxy.id":
		return r.ProxyID
	case "proxy.name":
		return r.ProxyName
	case "proxy.route.path":
		return r.ProxyRoutePath
	case "proxy.route.verb":
		return r.ProxyRouteVerb
	case "remote_endpoint.response.time":
		return r.RemoteEndpointResponseTime
	case "request.size":
		return r.RequestSize
	case "request.id":
		return r.RequestID
	case "response.time":
		return r.ResponseTime
	case "response.size":
		return r.ResponseSize
	case "response.status":
		return r.ResponseStatus
	case "response.error":
		return r.ResponseError
	default:
		return nil
	}
}
