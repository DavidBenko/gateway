UPDATE push_devices
SET
  name = ?,
  type = ?,
  token = ?,
  data = ?
WHERE push_devices.id = ?
  AND push_devices.remote_endpoint_id IN
    (SELECT remote_endpoints.id
      FROM remote_endpoints, apis, accounts
      WHERE remote_endpoints.api_id = apis.id
        AND apis.account_id = accounts.id
        AND accounts.id = ?);
