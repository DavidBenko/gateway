CREATE TABLE IF NOT EXISTS "remote_endpoint_environment_data_tmp" (
  "id" SERIAL PRIMARY KEY,
  "remote_endpoint_id" INTEGER NOT NULL,
  "environment_id" INTEGER NOT NULL,
  "data" JSONB NOT NULL,
  UNIQUE ("remote_endpoint_id", "environment_id"),
  FOREIGN KEY("remote_endpoint_id") REFERENCES "remote_endpoints"("id") ON DELETE CASCADE,
  FOREIGN KEY("environment_id") REFERENCES "environments"("id") ON DELETE CASCADE
);
INSERT INTO remote_endpoint_environment_data_tmp (
  remote_endpoint_id, environment_id, data
) SELECT remote_endpoint_id, environment_id, data FROM remote_endpoint_environment_data;
DROP TABLE remote_endpoint_environment_data;
ALTER TABLE remote_endpoint_environment_data_tmp RENAME TO remote_endpoint_environment_data
