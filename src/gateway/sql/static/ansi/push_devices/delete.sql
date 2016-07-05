DELETE FROM push_devices
WHERE push_devices.id = ?
  AND push_devices.remote_endpoint_id IN
    (SELECT remote_endpoints.id
      FROM remote_endpoints, apis, accounts
      WHERE remote_endpoints.id = ?
        AND remote_endpoints.api_id = apis.id
        AND apis.account_id = accounts.id
        AND accounts.id = ?);
