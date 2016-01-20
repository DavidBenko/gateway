SELECT pc.id AS id
  , cr.id AS proxy_endpoint_component_reference_id
  , pc.conditional AS conditional
  , pc.conditional_positive AS conditional_positive
  , pc.type AS type
  , pc.data AS data
  , pc.type_discriminator AS type_discriminator
FROM proxy_endpoint_components pc, proxy_endpoint_component_references cr
WHERE pc.id = cr.proxy_endpoint_component_id
  AND cr.proxy_endpoint_id = ?
ORDER BY cr.position ASC;
