SELECT
  COUNT(mqtt_sessions.id)
FROM mqtt_sessions, remote_endpoints, apis
WHERE mqtt_sessions.remote_endpoint_id = ?
  AND mqtt_sessions.type = ?
  AND mqtt_sessions.remote_endpoint_id = remote_endpoints.id
  AND remote_endpoints.api_id = ?
  AND remote_endpoints.api_id = apis.id
  AND apis.account_id = ?;
