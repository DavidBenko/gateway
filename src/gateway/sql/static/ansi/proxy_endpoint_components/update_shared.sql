UPDATE proxy_endpoint_components
  SET
    conditional = ?
    , conditional_positive = ?
    , type = ?
    , data = ?
    , name = ?
    , description = ?
  WHERE id = ?
    AND type_discriminator = 'shared'
    AND api_id = (SELECT id FROM apis WHERE id = ? AND account_id = ?)
