SELECT
  apis.id as api_id,
  job_tests.job_id as job_id,
  job_tests.id as id,
  job_tests.name as name,
  job_tests.parameters as parameters
FROM job_tests, proxy_endpoints, apis
WHERE job_tests.job_id = proxy_endpoints.id
  AND proxy_endpoints.api_id = ?
  AND proxy_endpoints.api_id = apis.id
  AND apis.account_id = ?
ORDER BY job_tests.id ASC;
