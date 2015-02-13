INSERT INTO hosts (api_id, name)
VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?)
