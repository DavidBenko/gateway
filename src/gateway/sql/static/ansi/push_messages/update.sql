UPDATE push_messages
SET
  stamp = ?,
  data = ?
WHERE push_messages.id = ?
  AND push_messages.push_device_id IN
    (SELECT push_devices.id
      FROM push_devices, remote_endpoints, apis
      WHERE push_devices.id = ?
        AND push_devices.remote_endpoint_id = remote_endpoints.id
        AND remote_endpoints.api_id = apis.id
        AND apis.account_id = ?)
  AND push_messages.push_channel_id IN
    (SELECT push_channels.id
      FROM push_channels
      WHERE push_channels.id = ?
      AND push_channels.account_id = ?)
