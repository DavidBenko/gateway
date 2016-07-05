SELECT
  push_channel_messages.id as id,
  push_channel_messages.push_channel_id as push_channel_id,
  push_channel_messages.stamp as stamp,
  push_channel_messages.data as data
FROM push_channel_messages, push_channels
WHERE push_channel_messages.push_channel_id = push_channels.id
  AND push_channels.account_id = ?
ORDER BY push_channel_messages.stamp DESC;
