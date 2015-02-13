UPDATE environments
SET name = ?, description = ?, data = ?
WHERE environments.id = ?
  AND environments.api_id IN
  (SELECT id FROM apis WHERE id = ? AND account_id = ?);
