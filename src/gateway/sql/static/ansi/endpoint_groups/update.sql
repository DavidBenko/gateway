UPDATE endpoint_groups
SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP
WHERE endpoint_groups.id = ?
  AND endpoint_groups.api_id IN
        (SELECT id FROM apis WHERE id = ? AND account_id = ?)
