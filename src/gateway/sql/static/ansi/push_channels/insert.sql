INSERT INTO push_channels (
  remote_endpoint_id,
  name,
  expires,
  data
)
VALUES (
  (SELECT remote_endpoints.id
    FROM remote_endpoints, apis
    WHERE remote_endpoints.id = ?
      AND remote_endpoints.api_id = ?
      AND remote_endpoints.api_id = apis.id
      AND apis.account_id = ?),
  ?,
  ?,
  ?
)
