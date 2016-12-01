ALTER TABLE proxy_endpoint_tests ADD COLUMN created_at TIMESTAMPTZ;
ALTER TABLE proxy_endpoint_tests ADD COLUMN updated_at TIMESTAMPTZ;
ALTER TABLE proxy_endpoint_tests ADD COLUMN channels BOOLEAN;
ALTER TABLE proxy_endpoint_tests ADD COLUMN channel_id INTEGER;
ALTER TABLE proxy_endpoint_tests ADD FOREIGN KEY("channel_id") REFERENCES "proxy_endpoint_channels"("id") ON DELETE CASCADE;
UPDATE proxy_endpoint_tests SET channels = FALSE;
ALTER TABLE proxy_endpoint_tests ALTER COLUMN channels SET NOT NULL;
