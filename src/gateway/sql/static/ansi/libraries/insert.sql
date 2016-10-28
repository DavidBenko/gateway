INSERT INTO libraries (api_id, name, description, data, created_at)
VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?, ?, ?, CURRENT_TIMESTAMP)
