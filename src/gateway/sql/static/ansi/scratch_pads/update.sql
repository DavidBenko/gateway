UPDATE scratch_pads
SET
  name = ?,
  code = ?,
  data = ?
WHERE scratch_pads.id = ?
  AND scratch_pads.environment_data_id IN
    (SELECT remote_endpoint_environment_data.id
      FROM remote_endpoint_environment_data, remote_endpoints, apis
      WHERE remote_endpoint_environment_data.id = ?
        AND remote_endpoint_environment_data.remote_endpoint_id = ?
        AND remote_endpoint_environment_data.remote_endpoint_id = remote_endpoints.id
        AND remote_endpoints.api_id = ?
        AND remote_endpoints.api_id = apis.id
        AND apis.account_id = ?);
