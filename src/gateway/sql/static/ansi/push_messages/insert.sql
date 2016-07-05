INSERT INTO push_messages (
  push_device_id,
  push_channel_id,
  push_channel_message_id,
  stamp,
  data
)
VALUES (
  (SELECT push_devices.id
    FROM push_devices, push_channels_push_devices, push_channels
    WHERE push_devices.id = ?
      AND push_devices.id = push_channels_push_devices.push_device_id
      AND push_channels_push_devices.push_channel_id = ?
      AND push_channels.id = push_channels_push_devices.push_channel_id
      AND push_channels.account_id = ?),
  (SELECT push_channels.id
    FROM push_channels
    WHERE push_channels.id = ?
      AND push_channels.account_id = ?),
  ?,
  ?,
  ?
)
