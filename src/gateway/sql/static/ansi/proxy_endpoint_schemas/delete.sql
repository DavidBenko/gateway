DELETE FROM proxy_endpoint_schemas
WHERE proxy_endpoint_schemas.id = ?
  AND proxy_endpoint_schemas.endpoint_id IN
    (SELECT proxy_endpoints.id
      FROM proxy_endpoints, apis
      WHERE proxy_endpoints.id = ?
        AND proxy_endpoints.api_id = ?
        AND proxy_endpoints.api_id = apis.id
        AND apis.account_id = ?);
