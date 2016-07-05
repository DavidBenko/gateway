INSERT INTO push_channel_messages (
  push_channel_id,
  stamp,
  data
)
VALUES ((SELECT push_channels.id
  FROM push_channels
  WHERE push_channels.id = ?
    AND push_channels.account_id = ?),
?,
?)
