UPDATE libraries
SET name = ?, description = ?, data = ?, updated_at = CURRENT_TIMESTAMP
WHERE libraries.id = ?
  AND libraries.api_id IN
  (SELECT id FROM apis WHERE id = ? AND account_id = ?);
