CREATE TABLE IF NOT EXISTS "mqtt_sessions" (
  "id" SERIAL PRIMARY KEY,
  "remote_endpoint_id" INTEGER NOT NULL,
  "type" TEXT NOT NULL,
  "client_id" TEXT NOT NULL,
  "data" JSONB NOT NULL,
  UNIQUE ("remote_endpoint_id", "type", "client_id"),
  FOREIGN KEY("remote_endpoint_id") REFERENCES "remote_endpoints"("id") ON DELETE CASCADE
);
CREATE INDEX idx_mqtt_sessions_type ON mqtt_sessions USING btree(type);
CREATE INDEX idx_mqtt_sessions_client_id ON mqtt_sessions USING btree(client_id);
