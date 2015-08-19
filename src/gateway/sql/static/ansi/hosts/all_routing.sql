SELECT 
  hosts.id as id,
  hosts.hostname as hostname,
  hosts.api_id as api_id,
  apis.account_id as account_id
FROM hosts, apis
WHERE hosts.api_id = apis.id;
