UPDATE shared_components
SET
  -- ProxyEndpointComponent
  conditional = ?,
  conditional_positive = ?,
  type = ?,
  data = ?,
  -- SharedComponent
  api_id = ?,
  name = ?,
  description = ?
WHERE shared_components.id = ?
AND shared_components.api_id IN (
  SELECT id
    FROM apis
    WHERE id = ?
    AND account_id = ?
  );
