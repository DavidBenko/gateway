DELETE FROM push_messages
WHERE push_messages.id = ?
  AND push_messages.push_device_id IN
    (SELECT push_devices.id
      FROM push_devices, push_channels, remote_endpoints, apis
      WHERE push_devices.id = ?
        AND push_devices.push_channel_id = ?
        AND push_devices.push_channel_id = push_channels.id
        AND push_channels.remote_endpoint_id = ?
        AND push_channels.remote_endpoint_id = remote_endpoints.id
        AND remote_endpoints.api_id = ?
        AND remote_endpoints.api_id = apis.id
        AND apis.account_id = ?);
