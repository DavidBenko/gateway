SELECT
  api_id,
  id,
  name,
  description,
  data
FROM environments
WHERE id = ?;
