UPDATE push_devices
SET
  name = ?,
  type = ?,
  token = ?,
  data = ?
WHERE push_devices.id = ?
  AND push_devices.id IN
    (SELECT push_channels_push_devices.push_device_id
      FROM push_channels_push_devices, push_channels
      WHERE push_channels_push_devices.push_channel_id = ?
        AND push_channels_push_devices.push_channel_id = push_channels.id
        AND push_channels.account_id = ?);
