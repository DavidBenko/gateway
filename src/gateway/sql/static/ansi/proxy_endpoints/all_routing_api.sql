SELECT id, routes, cors_enabled
FROM proxy_endpoints
WHERE active = ?
  AND api_id = ?;
