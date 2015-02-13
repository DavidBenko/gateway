SELECT
  hosts.api_id as api_id,
  hosts.id as id,
  hosts.name as name
FROM hosts, apis
WHERE hosts.id = ?
  AND hosts.api_id = ?
  AND hosts.api_id = apis.id
  AND apis.account_id = ?;
