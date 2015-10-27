INSERT INTO shared_components (
  -- ProxyEndpointComponent
  conditional,
  conditional_positive,
  type,
  data,
  -- SharedComponent
  api_id,
  name,
  description,
) VALUES (
  ?, ?, ?, ?,
  (SELECT id
    FROM apis
    WHERE id = ? AND account_id = ?
  ),
  ?, ?
)
