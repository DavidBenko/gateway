DELETE FROM mqtt_sessions
WHERE mqtt_sessions.type = ?
  AND mqtt_sessions.client_id = ?
  AND mqtt_sessions.remote_endpoint_id IN
    (SELECT remote_endpoints.id
      FROM remote_endpoints, apis
      WHERE remote_endpoints.id = ?
        AND remote_endpoints.api_id = ?
        AND remote_endpoints.api_id = apis.id
        AND apis.account_id = ?);
