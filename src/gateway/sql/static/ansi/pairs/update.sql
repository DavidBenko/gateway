UPDATE proxy_endpoint_test_pairs
SET
  type = ?,
  key = ?,
  value = ?
WHERE id = ? AND test_id = ?;
