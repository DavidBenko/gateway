SELECT
  id, account_id, name, key
FROM keys
WHERE name = ?
  AND account_id = ?
