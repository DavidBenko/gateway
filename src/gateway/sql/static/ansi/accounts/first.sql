SELECT id, name, plan_id, stripe_customer_id, stripe_subscription_id, stripe_payment_retry_attempt
FROM accounts
ORDER BY id
LIMIT 1;
