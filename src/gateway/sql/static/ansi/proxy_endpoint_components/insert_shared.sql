INSERT INTO proxy_endpoint_components (
  conditional
  , conditional_positive
  , type
  , data
  , name
  , description
  , type_discriminator
  , api_id
) VALUES (
  ?, ?, ?, ?
  , ?, ?
  , 'shared'
  , (SELECT id FROM apis WHERE id = ? AND account_id = ?)
)
