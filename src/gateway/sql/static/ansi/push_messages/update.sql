UPDATE push_messages
SET
  stamp = ?,
  data = ?
WHERE push_messages.id = ?
  AND push_messages.push_device_id IN
    (SELECT push_devices.id
      FROM push_devices, push_channels
      WHERE push_devices.id = ?
        AND push_devices.push_channel_id = ?
        AND push_devices.push_channel_id = push_channels.id
        AND push_channels.account_id = ?);
