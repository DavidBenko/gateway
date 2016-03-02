CREATE TABLE IF NOT EXISTS "scratch_pads" (
  "id" SERIAL PRIMARY KEY,
  "remote_endpoint_environment_data_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "code" TEXT NOT NULL,
  "data" JSONB NOT NULL,
  UNIQUE ("remote_endpoint_environment_data_id", "name"),
  FOREIGN KEY("remote_endpoint_environment_data_id") REFERENCES "remote_endpoint_environment_data"("id") ON DELETE CASCADE
);
CREATE INDEX idx_scratch_pads_remote_endpoint_environment_data_id ON scratch_pads USING btree(remote_endpoint_environment_data_id);
