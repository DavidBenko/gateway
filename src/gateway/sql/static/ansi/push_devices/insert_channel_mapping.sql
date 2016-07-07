INSERT INTO push_channels_push_devices (
  push_device_id,
  push_channel_id,
  expires
)
VALUES (
  ?,
  (SELECT push_channels.id
    FROM push_channels
    WHERE push_channels.id = ?
      AND push_channels.account_id = ?),
  ?
)
