SELECT
  api_id,
  id,
  name,
  description,
  endpoint_group_id,
  environment_id,
  active,
  cors_enabled,
  routes
FROM proxy_endpoints
WHERE id = ?
  AND type = ?;
