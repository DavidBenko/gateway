SELECT
  pc.id as id
  , pc.conditional as conditional
  , pc.conditional_positive as conditional_positive
  , pc.type as type
  , pc.data as data
  , pc.api_id as api_id
  , pc.name as name
  , pc.description as description
FROM proxy_endpoint_components AS pc, apis
WHERE pc.api_id = apis.id
  AND pc.type_discriminator = 'shared'
  AND pc.id = ?
  AND pc.api_id = ?
  AND apis.account_id = ?;
