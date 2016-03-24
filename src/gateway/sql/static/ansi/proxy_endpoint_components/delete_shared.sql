DELETE FROM proxy_endpoint_components
WHERE type_discriminator = 'shared'
  AND id = ?
  AND api_id = (SELECT id FROM apis WHERE id = ? AND account_id = ?);
