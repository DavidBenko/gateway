SELECT
  proxy_endpoints.api_id as api_id,
  proxy_endpoints.id as id,
  proxy_endpoints.type as type,
  proxy_endpoints.name as name,
  proxy_endpoints.description as description,
  proxy_endpoints.endpoint_group_id as endpoint_group_id,
  proxy_endpoints.environment_id as environment_id,
  proxy_endpoints.active as active
FROM proxy_endpoints, apis
WHERE proxy_endpoints.api_id = ?
  AND proxy_endpoints.api_id = apis.id
  AND apis.account_id = ?
ORDER BY
  proxy_endpoints.name ASC,
  proxy_endpoints.id ASC;
