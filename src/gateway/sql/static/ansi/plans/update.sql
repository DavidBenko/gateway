UPDATE plans
SET name = ?, stripe_name = ?, max_users = ?, javascript_timeout = ?, job_timeout = ?, price = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
