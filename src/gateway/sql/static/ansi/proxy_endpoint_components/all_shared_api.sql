SELECT
  pc.id as id
  , pc.conditional as conditional
  , pc.conditional_positive as conditional_positive
  , pc.type as type
  , pc.data as data
  , pc.api_id as api_id
  , pc.name as name
  , pc.description as description
FROM proxy_endpoint_components pc
WHERE pc.api_id = ?
  AND pc.type_discriminator = 'shared'
ORDER BY pc.name ASC;
