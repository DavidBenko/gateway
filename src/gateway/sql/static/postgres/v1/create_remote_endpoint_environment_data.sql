CREATE TABLE IF NOT EXISTS "remote_endpoint_environment_data" (
  "remote_endpoint_id" INTEGER NOT NULL,
  "environment_id" INTEGER NOT NULL,
  "data" TEXT NOT NULL,
  PRIMARY KEY("remote_endpoint_id", "environment_id"),
  FOREIGN KEY("remote_endpoint_id") REFERENCES "remote_endpoints"("id") ON DELETE CASCADE,
  FOREIGN KEY("environment_id") REFERENCES "environments"("id") ON DELETE CASCADE
);
