SELECT
  custom_functions.api_id as api_id,
  custom_function_files.custom_function_id as custom_function_id,
  custom_function_files.id as id,
  custom_function_files.name as name,
  custom_function_files.body as body
FROM custom_function_files, custom_functions, apis
WHERE custom_function_files.id = ?
  AND custom_function_files.custom_function_id = ?
  AND custom_function_files.custom_function_id = custom_functions.id
  AND custom_functions.api_id = ?
  AND custom_functions.api_id = apis.id
  AND apis.account_id = ?;
