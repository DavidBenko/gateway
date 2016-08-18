UPDATE accounts
SET stripe_payment_retry_attempt = ?
WHERE id = ?;
