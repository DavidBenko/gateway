SELECT id, name, plan_id, stripe_customer_id, stripe_subscription_id, stripe_payment_retry_attempt
FROM accounts
WHERE id = ?;
