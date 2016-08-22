SELECT id, name, stripe_name, max_users, javascript_timeout, price
FROM plans
WHERE name = ?;
