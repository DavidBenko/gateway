SELECT id, api_id, routes
FROM proxy_endpoints
WHERE active = ?;
