UPDATE proxy_endpoint_tests
SET
  name = ?,
  methods = ?,
  route = ?,
  body = ?,
  data = ?
WHERE id = ? AND endpoint_id = ?;
