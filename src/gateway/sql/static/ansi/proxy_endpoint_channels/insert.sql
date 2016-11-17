INSERT INTO proxy_endpoint_channels (
  proxy_endpoint_id,
  remote_endpoint_id,
  name,
  created_at
)
VALUES (
  (SELECT proxy_endpoints.id
    FROM proxy_endpoints, apis
    WHERE proxy_endpoints.id = ?
      AND proxy_endpoints.api_id = ?
      AND proxy_endpoints.api_id = apis.id
      AND apis.account_id = ?),
  (SELECT remote_endpoints.id
    FROM remote_endpoints, apis
    WHERE remote_endpoints.id = ?
      AND remote_endpoints.api_id = ?
      AND remote_endpoints.api_id = apis.id
      AND apis.account_id = ?),
  ?,
  CURRENT_TIMESTAMP
)
