SELECT
  push_devices.id as id,
  push_devices.push_channel_id as push_channel_id,
  push_devices.name as name,
  push_devices.type as type,
  push_devices.token as token,
  push_devices.expires as expires,
  push_devices.data as data
FROM push_devices, push_channels, remote_endpoints, apis
WHERE push_devices.id = ?
  AND push_devices.push_channel_id = ?
  AND push_devices.push_channel_id = push_channels.id
  AND push_channels.remote_endpoint_id = ?
  AND push_channels.remote_endpoint_id = remote_endpoints.id
  AND remote_endpoints.api_id = ?
  AND remote_endpoints.api_id = apis.id
  AND apis.account_id = ?;
