UPDATE sessions
SET
  max_age = ?,
  expires = ?,
  session_values = ?
WHERE sessions.id = ?
  AND sessions.session_name = ?
  AND sessions.session_uuid = ?;
