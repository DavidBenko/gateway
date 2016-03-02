UPDATE environments
SET
  name = ?,
  description = ?,
  data = ?,
  session_type = ?,
  session_header = ?,
  session_name = ?,
  session_auth_key = ?,
  session_encryption_key = ?,
  session_auth_key_rotate = ?,
  session_encryption_key_rotate = ?,
  show_javascript_errors = ?
WHERE environments.id = ?
  AND environments.api_id IN
  (SELECT id FROM apis WHERE id = ? AND account_id = ?);
