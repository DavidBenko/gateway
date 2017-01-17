INSERT INTO custom_functions (
  api_id,
  name, description,
  active, memory,
  cpu_shares, created_at
)
VALUES (
  (SELECT id FROM apis WHERE id = ? AND account_id = ?),
  ?, ?,
  ?, ?,
  ?, CURRENT_TIMESTAMP
)
