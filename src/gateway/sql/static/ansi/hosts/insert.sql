INSERT INTO hosts (api_id, name, hostname, created_at)
VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?,?,CURRENT_TIMESTAMP)
