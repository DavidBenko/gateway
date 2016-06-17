SELECT
  push_devices.id as id,
  push_devices.remote_endpoint_id as remote_endpoint_id,
  push_devices.name as name,
  push_devices.type as type,
  push_devices.token as token,
  push_channels_push_devices.expires as expires,
  push_devices.data as data
FROM push_devices, push_channels_push_devices, push_channels
WHERE push_devices.token = ?
  AND push_channels_push_devices.push_device_id = push_devices.id
  AND push_channels_push_devices.push_channel_id = ?
  AND push_channels_push_devices.push_channel_id = push_channels.id
  AND push_channels.account_id = ?;
