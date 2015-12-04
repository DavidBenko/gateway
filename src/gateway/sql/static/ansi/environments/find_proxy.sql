SELECT
  api_id,
  id,
  name,
  description,
  data,
  session_name,
  session_auth_key,
  session_encryption_key,
  session_auth_key_rotate,
  session_encryption_key_rotate,
  show_javascript_errors
FROM environments
WHERE id = ?;
