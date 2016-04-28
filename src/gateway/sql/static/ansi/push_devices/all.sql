SELECT
  push_devices.id as id,
  push_devices.push_channel_id as push_channel_id,
  push_devices.name as name,
  push_devices.type as type,
  push_devices.token as token,
  push_devices.expires as expires,
  push_devices.data as data
FROM push_devices, push_channels
WHERE push_devices.push_channel_id = ?
  AND push_devices.push_channel_id = push_channels.id
  AND push_channels.account_id = ?
ORDER BY push_channels.id ASC;
