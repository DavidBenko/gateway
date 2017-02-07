UPDATE custom_function_tests
SET
  name = ?,
  input = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE custom_function_tests.id = ?
  AND custom_function_tests.custom_function_id IN
    (SELECT custom_functions.id
      FROM custom_functions, apis
      WHERE custom_functions.id = ?
        AND custom_functions.api_id = ?
        AND custom_functions.api_id = apis.id
        AND apis.account_id = ?);
