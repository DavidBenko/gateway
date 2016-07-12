SELECT
  push_channels.id as id,
  push_channels.account_id as account_id,
  push_channels.api_id as api_id,
  push_channels.remote_endpoint_id as remote_endpoint_id,
  push_channels.name as name,
  push_channels_push_devices.expires as expires,
  push_channels_push_devices.qos as qos,
  push_channels.data as data
FROM push_channels, push_channels_push_devices, push_devices
WHERE push_channels.account_id = ?
  AND push_channels.id = push_channels_push_devices.push_channel_id
  AND push_channels_push_devices.push_device_id = push_devices.id
  AND push_devices.token = ?
ORDER BY push_channels.id ASC;
