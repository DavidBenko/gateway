SELECT
  libraries.api_id as api_id,
  libraries.id as id,
  libraries.name as name,
  libraries.description as description,
  libraries.data as data
FROM libraries, apis
WHERE libraries.api_id = ?
  AND libraries.api_id = apis.id
  AND apis.account_id = ?
ORDER BY libraries.name ASC;
