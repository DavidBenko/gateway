SELECT id, name, plan_id, stripe_customer_id, stripe_subscription_id
FROM accounts
WHERE id = ?;
