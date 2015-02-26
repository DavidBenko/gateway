INSERT INTO hosts (api_id, name, hostname)
VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?,?)
