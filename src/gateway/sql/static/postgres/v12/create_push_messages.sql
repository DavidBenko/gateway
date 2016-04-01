CREATE TABLE IF NOT EXISTS "push_messages" (
  "id" SERIAL PRIMARY KEY,
  "push_device_id" INTEGER NOT NULL,
  "stamp" INTEGER NOT NULL,
  "data" JSONB NOT NULL,
  FOREIGN KEY("push_device_id") REFERENCES "push_devices"("id") ON DELETE CASCADE
);
CREATE INDEX idx_push_messages_push_device_id ON push_messages USING btree(push_device_id);
