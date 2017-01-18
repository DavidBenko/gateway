SELECT
  hosts.api_id as api_id,
  hosts.id as id,
  hosts.name as name,
  hosts.hostname as hostname
FROM hosts
WHERE hosts.hostname = ?
