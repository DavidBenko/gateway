DELETE FROM push_channels
WHERE push_channels.id = ?
  AND push_channels.account_id = ?;
