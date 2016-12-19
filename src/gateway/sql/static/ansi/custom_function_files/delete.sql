DELETE FROM custom_function_files
WHERE custom_function_files.id = ?
  AND custom_function_files.custom_function_id IN
    (SELECT custom_function.id
      FROM custom_function, apis
      WHERE custom_function.id = ?
        AND custom_function.api_id = ?
        AND custom_function.api_id = apis.id
        AND apis.account_id = ?);
