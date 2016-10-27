ALTER TABLE accounts ADD COLUMN created_at DATETIME;
ALTER TABLE accounts ADD COLUMN updated_at DATETIME;

ALTER TABLE apis ADD COLUMN created_at DATETIME;
ALTER TABLE apis ADD COLUMN updated_at DATETIME;

ALTER TABLE endpoint_groups ADD COLUMN created_at DATETIME;
ALTER TABLE endpoint_groups ADD COLUMN updated_at DATETIME;

ALTER TABLE environments ADD COLUMN created_at DATETIME;
ALTER TABLE environments ADD COLUMN updated_at DATETIME;

ALTER TABLE hosts ADD COLUMN created_at DATETIME;
ALTER TABLE hosts ADD COLUMN updated_at DATETIME;

ALTER TABLE keys ADD COLUMN created_at DATETIME;
ALTER TABLE keys ADD COLUMN updated_at DATETIME;

ALTER TABLE libraries ADD COLUMN created_at DATETIME;
ALTER TABLE libraries ADD COLUMN updated_at DATETIME;

ALTER TABLE plans ADD COLUMN created_at DATETIME;
ALTER TABLE plans ADD COLUMN updated_at DATETIME;

ALTER TABLE proxy_endpoints ADD COLUMN created_at DATETIME;
ALTER TABLE proxy_endpoints ADD COLUMN updated_at DATETIME;

ALTER TABLE remote_endpoints ADD COLUMN created_at DATETIME;
ALTER TABLE remote_endpoints ADD COLUMN updated_at DATETIME;

ALTER TABLE scratch_pads ADD COLUMN created_at DATETIME;
ALTER TABLE scratch_pads ADD COLUMN updated_at DATETIME;

ALTER TABLE timers ADD COLUMN created_at DATETIME;
ALTER TABLE timers ADD COLUMN updated_at DATETIME;

ALTER TABLE users ADD COLUMN created_at DATETIME;
ALTER TABLE users ADD COLUMN updated_at DATETIME;
