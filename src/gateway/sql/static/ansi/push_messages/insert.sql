INSERT INTO push_messages (
  push_device_id,
  stamp,
  data
)
VALUES (
  (SELECT push_devices.id
    FROM push_devices, push_channels
    WHERE push_devices.id = ?
      AND push_devices.push_channel_id = ?
      AND push_devices.push_channel_id = push_channels.id
      AND push_channels.account_id = ?),
  ?,
  ?
)
