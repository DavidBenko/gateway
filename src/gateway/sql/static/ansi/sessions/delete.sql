DELETE FROM sessions
WHERE sessions.id = ?
  AND sessions.session_name = ?
  AND sessions.session_uuid = ?;
