ALTER TABLE proxy_endpoints ADD COLUMN type TEXT NOT NULL DEFAULT '';
ALTER TABLE proxy_endpoints DROP CONSTRAINT proxy_endpoints_api_id_name_key;
ALTER TABLE proxy_endpoints ADD UNIQUE("api_id", "type", "name");
UPDATE proxy_endpoints SET type='http';
