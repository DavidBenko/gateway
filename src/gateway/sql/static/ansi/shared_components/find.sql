SELECT
  -- ProxyEndpointComponent
  shared_components.id as id,
  shared_components.conditional as conditional,
  shared_components.conditional_positive as conditional_positive,
  shared_components.type as type,
  shared_components.data as data,
  -- SharedComponent
  shared_components.api_id as api_id,
  shared_components.name as name,
  shared_components.description as description
FROM shared_components, apis
WHERE shared_components.id = ?
  AND shared_components.api_id = ?
  AND shared_components.api_id = apis.id
  AND apis.account_id = ?;
