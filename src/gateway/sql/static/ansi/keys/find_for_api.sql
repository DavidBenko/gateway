SELECT
  keys.id as id,
  keys.name as name,
  keys.api_id as api_id,
  keys.key as key 
FROM keys
WHERE keys.api_id = ?
ORDER BY keys.id ASC;
