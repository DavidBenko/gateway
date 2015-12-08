SELECT
  proxy_endpoint_schemas.endpoint_id as endpoint_id,
  proxy_endpoint_schemas.id as id,
  proxy_endpoint_schemas.name as name,
  proxy_endpoint_schemas.request_schema_id as request_schema_id,
  proxy_endpoint_schemas.request_type as request_type,
  proxy_endpoint_schemas.request_schema as request_schema,
  proxy_endpoint_schemas.response_same_as_request as response_same_as_request,
  proxy_endpoint_schemas.response_schema_id as response_schema_id,
  proxy_endpoint_schemas.response_type as response_type,
  proxy_endpoint_schemas.response_schema as response_schema,
  proxy_endpoint_schemas.data as data
FROM proxy_endpoint_schemas, proxy_endpoints, apis
WHERE proxy_endpoint_schemas.endpoint_id = ?
  AND proxy_endpoint_schemas.endpoint_id = proxy_endpoints.id
  AND proxy_endpoints.api_id = ?
  AND proxy_endpoints.api_id = apis.id
  AND apis.account_id = ?
ORDER BY proxy_endpoint_schemas.id ASC;
