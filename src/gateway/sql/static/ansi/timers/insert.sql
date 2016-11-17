INSERT INTO timers (
  api_id,
  job_id,
  name, once, time_zone,
  minute, hour, day_of_month, month, day_of_week,
  next, parameters, data, created_at
)
VALUES (
  (SELECT id FROM apis WHERE id = ? AND account_id = ?),
  (SELECT id FROM proxy_endpoints WHERE id = ? AND api_id = ?),
  ?, ?, ?,
  ?, ?, ?, ?, ?,
  ?, ?, ?, CURRENT_TIMESTAMP
)
