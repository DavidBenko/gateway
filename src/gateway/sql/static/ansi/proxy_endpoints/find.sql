SELECT
  proxy_endpoints.api_id as api_id,
  proxy_endpoints.id as id,
  proxy_endpoints.type as type,
  proxy_endpoints.name as name,
  proxy_endpoints.description as description,
  proxy_endpoints.endpoint_group_id as endpoint_group_id,
  proxy_endpoints.environment_id as environment_id,
  proxy_endpoints.active as active,
  proxy_endpoints.cors_enabled as cors_enabled,
  proxy_endpoints.routes as routes
FROM proxy_endpoints, apis
WHERE proxy_endpoints.id = ?
  AND proxy_endpoints.type = ?
  AND proxy_endpoints.api_id = ?
  AND proxy_endpoints.api_id = apis.id
  AND apis.account_id = ?;
