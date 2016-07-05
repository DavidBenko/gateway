UPDATE push_channels_push_devices
SET
  expires = ?
WHERE push_channels_push_devices.push_device_id = ?
  AND push_channels_push_devices.push_channel_id IN
    (SELECT push_channels.id
      FROM push_channels
      WHERE push_channels.id = ?
        AND push_channels.account_id = ?);
