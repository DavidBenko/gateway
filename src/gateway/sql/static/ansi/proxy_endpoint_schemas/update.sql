UPDATE proxy_endpoint_schemas
SET
  name = ?,
  request_schema_id =
    (SELECT schemas.id
      FROM schemas, apis
      WHERE schemas.id = ?
        AND schemas.api_id = ?
        AND schemas.api_id = apis.id
        AND apis.account_id = ?),
  request_type = ?,
  request_schema = ?,
  response_same_as_request = ?,
  response_schema_id =
    (SELECT schemas.id
      FROM schemas, apis
      WHERE schemas.id = ?
        AND schemas.api_id = ?
        AND schemas.api_id = apis.id
        AND apis.account_id = ?),
  response_type = ?,
  response_schema = ?,
  data = ?
WHERE proxy_endpoint_schemas.id = ?
  AND proxy_endpoint_schemas.endpoint_id IN
    (SELECT proxy_endpoints.id
      FROM proxy_endpoints, apis
      WHERE proxy_endpoints.id = ?
        AND proxy_endpoints.api_id = ?
        AND proxy_endpoints.api_id = apis.id
        AND apis.account_id = ?);
