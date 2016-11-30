SELECT
  id,
  name,
  channels,
  channel_id,
  methods,
  route,
  body,
  data
FROM proxy_endpoint_tests
WHERE endpoint_id = ?
ORDER BY
  name ASC,
  id ASC;
