SELECT pc.id AS id
  , cr_by_proxy_id.id AS proxy_endpoint_component_reference_id
  , pc.conditional AS conditional
  , pc.conditional_positive AS conditional_positive
  , pc.type AS type
  , pc.data AS data
  , pc.type_discriminator AS type_discriminator
FROM proxy_endpoint_components AS pc,
  (SELECT cr.id
    , cr.proxy_endpoint_component_id
    , cr.position
    FROM proxy_endpoint_component_references AS cr
    WHERE cr.proxy_endpoint_id = ?) AS cr_by_proxy_id
WHERE pc.id = cr_by_proxy_id.proxy_endpoint_component_id
ORDER BY cr_by_proxy_id.position ASC;
