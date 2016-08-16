SELECT
  mqtt_sessions.id as id,
  mqtt_sessions.remote_endpoint_id as remote_endpoint_id,
  mqtt_sessions.type as type,
  mqtt_sessions.client_id as client_id,
  mqtt_sessions.data as data
FROM mqtt_sessions, remote_endpoints, apis
WHERE mqtt_sessions.remote_endpoint_id = ?
  AND mqtt_sessions.type = ?
  AND mqtt_sessions.client_id = ?
  AND mqtt_sessions.remote_endpoint_id = remote_endpoints.id
  AND remote_endpoints.api_id = ?
  AND remote_endpoints.api_id = apis.id
  AND apis.account_id = ?;
