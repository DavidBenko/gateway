SELECT
  timers.id as id,
  timers.api_id as api_id,
  timers.job_id as job_id,
  timers.name as name,
  timers.once as once,
  timers.time_zone as time_zone,
  timers.minute as minute,
  timers.hour as hour,
  timers.day_of_month as day_of_month,
  timers.month as month,
  timers.day_of_week as day_of_week,
  timers.next as next,
  timers.attributes as attributes,
  timers.data as data
FROM timers, apis
WHERE timers.api_id = apis.id
  AND apis.account_id = ?
ORDER BY
  timers.name ASC,
  timers.id ASC;
