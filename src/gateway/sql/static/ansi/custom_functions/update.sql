UPDATE custom_functions
SET
  name = ?,
  description = ?,
  active = ?,
  memory = ?,
  cpu_shares = ?,
  timeout = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE custom_functions.id = ?
  AND custom_functions.api_id IN
      (SELECT id FROM apis WHERE id = ? AND account_id = ?);
