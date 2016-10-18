DELETE FROM timers
WHERE timers.id = ?
  AND timers.api_id IN
      (SELECT id FROM apis WHERE account_id = ?);
