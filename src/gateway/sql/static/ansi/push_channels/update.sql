UPDATE push_channels
SET
  api_id =
  (SELECT apis.id
    FROM apis
    WHERE apis.id = ?
      AND apis.account_id = ?),
  remote_endpoint_id =
  (SELECT remote_endpoints.id
    FROM remote_endpoints, apis
    WHERE remote_endpoints.id = ?
      AND remote_endpoints.api_id = ?
      AND remote_endpoints.api_id = apis.id
      AND apis.account_id = ?),
  name = ?,
  expires = ?,
  data = ?
WHERE push_channels.id = ?
  AND push_channels.account_id = ?;
