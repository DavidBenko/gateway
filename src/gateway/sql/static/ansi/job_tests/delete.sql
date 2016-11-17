DELETE FROM job_tests
WHERE job_tests.id = ?
  AND job_tests.job_id IN
    (SELECT proxy_endpoints.id
      FROM proxy_endpoints, apis
      WHERE proxy_endpoints.id = ?
        AND proxy_endpoints.api_id = ?
        AND proxy_endpoints.api_id = apis.id
        AND apis.account_id = ?);
