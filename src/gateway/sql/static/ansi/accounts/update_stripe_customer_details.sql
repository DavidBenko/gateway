UPDATE accounts
SET stripe_customer_id = ?, stripe_subscription_id = ?, plan_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
