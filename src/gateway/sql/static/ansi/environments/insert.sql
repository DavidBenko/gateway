INSERT INTO environments (
  api_id, name, description, data,
  session_type, session_header,
  session_name, session_auth_key, session_encryption_key,
  session_auth_key_rotate, session_encryption_key_rotate,
  show_javascript_errors, created_at
)
VALUES (
  (SELECT id FROM apis WHERE id = ? AND account_id = ?),?, ?, ?,
  ?, ?,
  ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
)
