SELECT
  apis.id as api_id,
  custom_function_tests.custom_function_id as custom_function_id,
  custom_function_tests.id as id,
  custom_function_tests.name as name,
  custom_function_tests.input as input
FROM custom_function_tests, custom_functions, apis
WHERE custom_function_tests.id = ?
  AND custom_function_tests.custom_function_id = ?
  AND custom_function_tests.custom_function_id = custom_functions.id
  AND custom_functions.api_id = ?
  AND custom_functions.api_id = apis.id
  AND apis.account_id = ?;
