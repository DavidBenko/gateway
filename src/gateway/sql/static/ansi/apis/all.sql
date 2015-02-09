SELECT id, name, description, cors_allow
FROM apis
WHERE account_id = ?
ORDER BY name ASC;
