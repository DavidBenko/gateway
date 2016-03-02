SELECT
  scratch_pads.id as id,
  scratch_pads.remote_endpoint_environment_data_id as remote_endpoint_environment_data_id,
  scratch_pads.name as name,
  scratch_pads.code as code,
  scratch_pads.data as data
FROM scratch_pads, remote_endpoint_environment_data, remote_endpoints, apis
WHERE scratch_pads.id = ?
  AND scratch_pads.remote_endpoint_environment_data_id = ?
  AND scratch_pads.remote_endpoint_environment_data_id = remote_endpoint_environment_data.id
  AND remote_endpoint_environment_data.remote_endpoint_id = ?
  AND remote_endpoint_environment_data.remote_endpoint_id = remote_endpoints.id
  AND remote_endpoints.api_id = ?
  AND remote_endpoints.api_id = apis.id
  AND apis.account_id = ?;
