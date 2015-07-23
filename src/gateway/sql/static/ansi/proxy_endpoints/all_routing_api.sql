SELECT id, routes, active, cors_enabled
FROM proxy_endpoints
WHERE api_id = ?;
