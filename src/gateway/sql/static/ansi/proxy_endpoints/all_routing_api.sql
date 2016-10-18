SELECT id, routes, active, cors_enabled
FROM proxy_endpoints
WHERE type = 'http' AND api_id = ?;
