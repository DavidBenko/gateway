SELECT
  api_id,
  id,
  name,
  description,
  data,
  session_auth_key,
  session_encryption_key,
  session_auth_key_rotate,
  session_encryption_key_rotate
FROM environments
WHERE id = ?;
