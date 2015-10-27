DELETE FROM shared_components
WHERE shared_components.id = ?
  AND shared_components.api_id IN
    (SELECT id FROM apis WHERE id = ? AND account_id = ?);
