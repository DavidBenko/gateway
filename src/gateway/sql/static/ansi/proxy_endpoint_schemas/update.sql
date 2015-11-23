UPDATE proxy_endpoint_schemas
SET
  name = ?,
  request_schema_id =
    (SELECT id FROM schemas WHERE id = ? AND api_id = ?),
  request_type = ?,
  request_schema = ?,
  response_same_as_request = ?,
  response_schema_id =
    (SELECT id FROM schemas WHERE id = ? AND api_id = ?),
  response_type = ?,
  response_schema = ?,
  data = ?
WHERE proxy_endpoint_schemas.id = ?
  AND proxy_endpoint_schemas.endpoint_id IN
    (SELECT id FROM proxy_endpoints WHERE id = ? AND api_id = ?);
