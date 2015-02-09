SELECT id, name, description, cors_allow
FROM apis
WHERE id = ?
  AND account_id = ?;
