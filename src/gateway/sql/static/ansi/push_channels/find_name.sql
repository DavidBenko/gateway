SELECT
  push_channels.id as id,
  push_channels.remote_endpoint_id as remote_endpoint_id,
  push_channels.name as name,
  push_channels.expires as expires,
  push_channels.data as data
FROM push_channels, remote_endpoints, apis
WHERE push_channels.name = ?
  AND push_channels.remote_endpoint_id = ?
  AND push_channels.remote_endpoint_id = remote_endpoints.id
  AND remote_endpoints.api_id = ?
  AND remote_endpoints.api_id = apis.id
  AND apis.account_id = ?;
