UPDATE apis
  SET name = ?,
      description = ?,
      cors_allow = ?
WHERE id = ?
  AND account_id = ?;
