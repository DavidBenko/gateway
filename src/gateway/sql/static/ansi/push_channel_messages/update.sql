UPDATE push_channel_messages
SET
  stamp = ?,
  data = ?
WHERE push_channel_messages.id = ?
  AND push_channel_messages.push_channel_id IN
    (SELECT push_channels.id
      FROM push_channels
      WHERE push_channels.id = ?
        AND push_channels.account_id = ?);
