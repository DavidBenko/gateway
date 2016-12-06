SELECT
  apis.id as api_id,
  proxy_endpoint_channels.proxy_endpoint_id as proxy_endpoint_id,
  proxy_endpoint_channels.id as id,
  proxy_endpoint_channels.remote_endpoint_id as remote_endpoint_id,
  proxy_endpoint_channels.name as name
FROM proxy_endpoint_channels, proxy_endpoints, apis
WHERE proxy_endpoint_channels.proxy_endpoint_id = proxy_endpoints.id
  AND proxy_endpoints.api_id = ?
  AND proxy_endpoints.api_id = apis.id
  AND apis.account_id = ?
ORDER BY proxy_endpoint_channels.id ASC;
