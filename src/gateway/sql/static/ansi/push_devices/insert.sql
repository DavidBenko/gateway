INSERT INTO push_devices (
  remote_endpoint_id,
  name,
  type,
  token,
  data
)
VALUES (
  (SELECT push_channels.remote_endpoint_id
    FROM push_channels
    WHERE push_channels.id = ?
      AND push_channels.account_id = ?),
  ?,
  ?,
  ?,
  ?
)
