INSERT INTO custom_function_files (
  custom_function_id,
  name, body,
  created_at
)
VALUES (
  (SELECT custom_functions.id
    FROM custom_functions, apis
    WHERE custom_functions.id = ?
      AND custom_functions.api_id = ?
      AND custom_functions.api_id = apis.id
      AND apis.account_id = ?),
  ?, ?,
  CURRENT_TIMESTAMP
)
