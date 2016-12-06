UPDATE proxy_endpoint_tests
SET
  name = ?,
  channels = ?,
  channel_id = ?,
  methods = ?,
  route = ?,
  body = ?,
  data = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND endpoint_id = ?;
