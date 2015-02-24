SELECT
  id, name, description, 
  cors_allow_origin, cors_allow_headers, cors_allow_credentials,
  cors_request_headers, cors_max_age
FROM apis
WHERE id = ?
  AND account_id = ?;
