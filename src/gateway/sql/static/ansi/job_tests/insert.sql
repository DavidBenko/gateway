INSERT INTO job_tests (
  job_id,
  name,
  parameters,
  created_at
)
VALUES (
  (SELECT proxy_endpoints.id
    FROM proxy_endpoints, apis
    WHERE proxy_endpoints.id = ?
      AND proxy_endpoints.api_id = ?
      AND proxy_endpoints.api_id = apis.id
      AND apis.account_id = ?),
  ?,
  ?,
  CURRENT_TIMESTAMP
)
