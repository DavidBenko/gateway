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
  (SELECT id FROM proxy_endpoints WHERE id = ? AND api_id = ?),
  ?,
  (SELECT id FROM schemas WHERE id = ? AND api_id = ?),
  ?,
  ?,
  ?,
  (SELECT id from schemas WHERE id = ? AND api_id = ?),
  ?,
  ?,
  ?
)
