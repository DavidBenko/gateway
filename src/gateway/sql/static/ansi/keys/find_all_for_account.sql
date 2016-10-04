SELECT
  keys.id as id,
  keys.account_id as account_id,
  keys.name as name,
  keys.key as key 
FROM keys
WHERE keys.account_id = ?
ORDER BY keys.id ASC;
