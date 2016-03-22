INSERT INTO proxy_endpoint_schemas (
  endpoint_id,
  name,
  request_schema_id,
  request_type,
  request_schema,
  response_same_as_request,
  response_schema_id,
  response_type,
  response_schema,
  data
)
VALUES (
  (SELECT proxy_endpoints.id
    FROM proxy_endpoints, apis
    WHERE proxy_endpoints.id = ?
      AND proxy_endpoints.api_id = ?
      AND proxy_endpoints.api_id = apis.id
      AND apis.account_id = ?),
  ?,
  (SELECT schemas.id
    FROM schemas, apis
    WHERE schemas.id = ?
      AND schemas.api_id = ?
      AND schemas.api_id = apis.id
      AND apis.account_id = ?),
  ?,
  ?,
  ?,
  (SELECT schemas.id
    FROM schemas, apis
    WHERE schemas.id = ?
      AND schemas.api_id = ?
      AND schemas.api_id = apis.id
      AND apis.account_id = ?),
  ?,
  ?,
  ?
)
