WITH q AS (
  SELECT
    pc.id AS id
    , cr.id AS proxy_endpoint_component_reference_id
    , pc.conditional AS conditional
    , pc.conditional_positive AS conditional_positive
    , pc.type AS type
    , pc.data AS data
    , pc.type_discriminator AS type_discriminator
    , cr.position AS position
  FROM proxy_endpoint_components pc, proxy_endpoint_component_references cr
  WHERE pc.id = cr.proxy_endpoint_component_id
    AND pc.type_discriminator = 'standard'
    AND cr.proxy_endpoint_id = ?
UNION
  SELECT
    pc.id AS id
    , cr.id AS proxy_endpoint_component_reference_id
    , '' AS conditional
    , FALSE AS conditional_positive
    , '' AS type
    , '{}' AS data
    , pc.type_discriminator AS type_discriminator
    , cr.position AS position
  FROM proxy_endpoint_components pc, proxy_endpoint_component_references cr
  WHERE pc.id = cr.proxy_endpoint_component_id
    AND pc.type_discriminator = 'shared'
    AND cr.proxy_endpoint_id = ?
)

SELECT
  id
  , proxy_endpoint_component_reference_id
  , conditional
  , conditional_positive
  , type
  , data
  , type_discriminator
FROM q
ORDER BY q.position ASC;
