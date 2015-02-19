SELECT
  api_id,
  id,
  name,
  description,
  endpoint_group_id,
  environment_id,
  active,
  cors_enabled,
  cors_allow_override,
  routes
FROM proxy_endpoints
WHERE id = ?;
