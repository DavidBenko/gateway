DELETE FROM proxy_endpoints
WHERE proxy_endpoints.id = ?
  AND proxy_endpoints.type = ?
  AND proxy_endpoints.api_id IN
      (SELECT id FROM apis WHERE id = ? AND account_id = ?);
