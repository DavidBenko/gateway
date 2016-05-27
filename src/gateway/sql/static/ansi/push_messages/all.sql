SELECT
  push_messages.id as id,
  push_messages.push_device_id as push_device_id,
  push_messages.stamp as stamp,
  push_messages.data as data
FROM push_messages, push_devices, push_channels
WHERE push_messages.push_device_id = ?
  AND push_messages.push_device_id = push_devices.id
  AND push_devices.push_channel_id = ?
  AND push_devices.push_channel_id = push_channels.id
  AND push_channels.account_id = ?
ORDER BY push_channels.id ASC;
