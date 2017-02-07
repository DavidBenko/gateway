DELETE FROM custom_functions
WHERE custom_functions.id = ?
  AND custom_functions.api_id IN
      (SELECT id FROM apis WHERE id = ? AND account_id = ?);
