INSERT INTO endpoint_groups
  (api_id, name, description)
VALUES
  ((SELECT id FROM apis WHERE id = ? AND account_id = ?), ?, ?)
