SELECT
  id,
  session_name,
  session_uuid,
  max_age,
  expires,
  session_values
FROM sessions
WHERE session_name = ?
  AND session_uuid = ?;
