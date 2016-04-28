SELECT
  push_channels.id as id,
  push_channels.account_id as account_id,
  push_channels.api_id as api_id,
  push_channels.remote_endpoint_id as remote_endpoint_id,
  push_channels.name as name,
  push_channels.expires as expires,
  push_channels.data as data
FROM push_channels
WHERE push_channels.account_id = ?
ORDER BY push_channels.id ASC;
