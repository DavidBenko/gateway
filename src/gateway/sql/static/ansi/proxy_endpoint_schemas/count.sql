SELECT
  COUNT(proxy_endpoint_schemas.id)
FROM proxy_endpoint_schemas, proxy_endpoints, apis
WHERE proxy_endpoint_schemas.endpoint_id = ?
  AND proxy_endpoint_schemas.endpoint_id = proxy_endpoints.id
  AND proxy_endpoints.api_id = ?
  AND proxy_endpoints.api_id = apis.id
  AND apis.account_id = ?;
