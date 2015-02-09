SELECT
  endpoint_groups.id as id,
  endpoint_groups.name as name,
  endpoint_groups.description as description
FROM endpoint_groups, apis
WHERE endpoint_groups.api_id = ?
  AND endpoint_groups.api_id = apis.id
  AND apis.account_id = ?
ORDER BY endpoint_groups.name ASC;
