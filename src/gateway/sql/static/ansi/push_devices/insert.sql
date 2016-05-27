INSERT INTO push_devices (
  push_channel_id,
  name,
  type,
  token,
  expires,
  data
)
VALUES (
  (SELECT push_channels.id
    FROM push_channels
    WHERE push_channels.id = ?
      AND push_channels.account_id = ?),
  ?,
  ?,
  ?,
  ?,
  ?
)
