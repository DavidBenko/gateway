UPDATE accounts
SET stripe_customer_id = ?, stripe_subscription_id = ?
WHERE id = ?;
