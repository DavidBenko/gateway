SELECT
  api_id,
  id,
  name,
  description,
  data
FROM libraries
WHERE api_id = ?
ORDER BY name ASC;
