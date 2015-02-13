INSERT INTO libraries (api_id, name, description, data)
VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?, ?, ?)
