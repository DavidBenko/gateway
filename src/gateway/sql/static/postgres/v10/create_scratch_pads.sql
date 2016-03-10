CREATE TABLE IF NOT EXISTS "scratch_pads" (
  "id" SERIAL PRIMARY KEY,
  "environment_data_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "code" TEXT NOT NULL,
  "data" JSONB NOT NULL,
  UNIQUE ("environment_data_id", "name"),
  FOREIGN KEY("environment_data_id") REFERENCES "remote_endpoint_environment_data"("id") ON DELETE CASCADE
);
CREATE INDEX idx_scratch_pads_environment_data_id ON scratch_pads USING btree(environment_data_id);
