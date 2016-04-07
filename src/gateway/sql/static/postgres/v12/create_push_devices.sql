CREATE TABLE IF NOT EXISTS "push_devices" (
  "id" SERIAL PRIMARY KEY,
  "push_channel_id" INTEGER NOT NULL,
  "environment_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "type" TEXT NOT NULL,
  "token" TEXT NOT NULL,
  "expires" INTEGER NOT NULL,
  "data" JSONB NOT NULL,
  UNIQUE ("push_channel_id", "name"),
  FOREIGN KEY("environment_id") REFERENCES "environments"("id") ON DELETE CASCADE,
  FOREIGN KEY("push_channel_id") REFERENCES "push_channels"("id") ON DELETE CASCADE
);
CREATE INDEX idx_push_devices_token ON push_devices USING btree(token);
