SELECT id, api_id, routes, cors_enabled
FROM proxy_endpoints
WHERE active = ?;
