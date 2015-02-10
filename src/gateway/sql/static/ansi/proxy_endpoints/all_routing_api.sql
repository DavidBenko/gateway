SELECT id, routes
FROM proxy_endpoints
WHERE active = ?
  AND api_id = ?;
