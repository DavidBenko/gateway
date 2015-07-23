SELECT
  id,
  key,
  value
FROM proxy_endpoint_test_pairs
WHERE test_id = ?
ORDER BY id ASC;
