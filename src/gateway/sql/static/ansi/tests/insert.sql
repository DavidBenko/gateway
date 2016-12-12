INSERT INTO proxy_endpoint_tests (
  endpoint_id, name,
  channels, channel_id,
  methods, route,
  body, data,
  created_at
)
VALUES (
  ?, ?,
  ?, ?,
  ?, ?,
  ?, ?,
  CURRENT_TIMESTAMP
)
