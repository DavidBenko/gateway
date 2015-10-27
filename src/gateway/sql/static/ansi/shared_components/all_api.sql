SELECT
  -- ProxyEndpointComponent
  id,
  conditional,
  conditional_positive,
  type,
  data,
  -- SharedComponent
  api_id,
  name,
  description
FROM shared_components
WHERE api_id = ?
ORDER BY name ASC;
