INSERT INTO sessions (
  session_name, session_uuid,
  max_age, expires, session_values
)
VALUES (
  ?, ?,
  ?, ?, ?
)
