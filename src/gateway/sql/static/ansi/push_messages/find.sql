SELECT
  push_messages.id as id,
  push_messages.push_device_id as push_device_id,
  push_messages.stamp as stamp,
  push_messages.data as data
FROM push_messages, push_devices, push_channels, remote_endpoints, apis
WHERE push_messages.id = ?
  AND push_messages.push_device_id = ?
  and push_messages.push_device_id = push_devices.id
  AND push_devices.push_channel_id = ?
  AND push_devices.push_channel_id = push_channels.id
  AND push_channels.remote_endpoint_id = ?
  AND push_channels.remote_endpoint_id = remote_endpoints.id
  AND remote_endpoints.api_id = ?
  AND remote_endpoints.api_id = apis.id
  AND apis.account_id = ?
ORDER BY push_messages.id;
