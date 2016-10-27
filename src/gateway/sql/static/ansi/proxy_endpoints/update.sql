UPDATE proxy_endpoints
SET
  name = ?,
  description = ?,
  endpoint_group_id =
    (SELECT id FROM endpoint_groups WHERE id = ? AND api_id = ?),
  environment_id =
    (SELECT id FROM environments WHERE id = ? AND api_id = ?),
  active = ?,
  cors_enabled = ?,
  routes = ?,
  updated_at = ?
WHERE proxy_endpoints.id = ?
  AND proxy_endpoints.type = ?
  AND proxy_endpoints.api_id IN
      (SELECT id FROM apis WHERE id = ? AND account_id = ?);
