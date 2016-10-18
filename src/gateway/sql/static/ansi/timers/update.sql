UPDATE timers
SET
  api_id =
    (SELECT id FROM apis WHERE id = ? AND account_id = ?),
  job_id =
    (SELECT id FROM proxy_endpoints WHERE id = ? AND api_id = ?),
  name = ?,
  once = ?,
  time_zone = ?,
  minute = ?,
  hour = ?,
  day_of_month = ?,
  month = ?,
  day_of_week = ?,
  next = ?,
  parameters = ?,
  data = ?
WHERE timers.id = ?
  AND timers.api_id IN
      (SELECT id FROM apis WHERE id = ? AND account_id = ?);
