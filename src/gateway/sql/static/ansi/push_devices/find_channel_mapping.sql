SELECT
  push_channels_push_devices.id as id,
  push_channels_push_devices.push_device_id as push_device_id,
  push_channels_push_devices.push_channel_id as push_channel_id,
  push_channels_push_devices.expires as expires
FROM push_channels_push_devices
WHERE push_channels_push_devices.push_device_id = ?
  AND push_channels_push_devices.push_channel_id IN
    (SELECT push_channels.id
      FROM push_channels
      WHERE push_channels.id = ?
        AND push_channels.account_id = ?);
