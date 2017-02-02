SELECT
  custom_functions.api_id as api_id,
  custom_functions.id as id,
  custom_functions.name as name,
  custom_functions.description as description,
  custom_functions.active as active,
  custom_functions.memory as memory,
  custom_functions.cpu_shares as cpu_shares,
  custom_functions.timeout as timeout
FROM custom_functions, apis
WHERE custom_functions.id = ?
  AND custom_functions.api_id = ?
  AND custom_functions.api_id = apis.id
  AND apis.account_id = ?;
