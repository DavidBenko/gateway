UPDATE hosts
SET name = ?
WHERE hosts.id = ?
  AND hosts.api_id IN
    (SELECT id FROM apis WHERE id = ? AND account_id = ?)
