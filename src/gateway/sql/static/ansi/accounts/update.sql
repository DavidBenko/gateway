UPDATE accounts
SET name = ?, plan_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
