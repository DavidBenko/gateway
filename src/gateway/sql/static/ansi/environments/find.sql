SELECT
  environments.api_id as api_id,
  environments.id as id,
  environments.name as name,
  environments.description as description,
  environments.data as data
FROM environments, apis
WHERE environments.id = ?
  AND environments.api_id = ?
  AND environments.api_id = apis.id
  AND apis.account_id = ?;
