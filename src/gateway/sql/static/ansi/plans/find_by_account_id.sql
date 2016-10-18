SELECT id, name, stripe_name, max_users, javascript_timeout, job_timeout, price
FROM plans
WHERE id = (SELECT plan_id from accounts where id = ?);
