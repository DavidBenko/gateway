DELETE FROM push_channels
WHERE push_channels.id = ?
  AND push_channels.remote_endpoint_id IN
    (SELECT remote_endpoints.id
      FROM remote_endpoints, apis
      WHERE remote_endpoints.id = ?
        AND remote_endpoints.api_id = ?
        AND remote_endpoints.api_id = apis.id
        AND apis.account_id = ?);
