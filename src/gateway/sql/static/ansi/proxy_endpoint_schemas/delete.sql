DELETE FROM proxy_endpoint_schemas
WHERE proxy_endpoint_schemas.id = ?
  AND proxy_endpoint_schemas.endpoint_id IN
    (SELECT id FROM proxy_endpoints WHERE id = ? AND api_id = ?);
