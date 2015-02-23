CREATE TABLE IF NOT EXISTS "proxy_endpoints" (
  "id" SERIAL PRIMARY KEY,
  "api_id" INTEGER NOT NULL,
  "endpoint_group_id" INTEGER,
  "environment_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "description" TEXT,
  "active" BOOLEAN NOT NULL DEFAULT TRUE,
  "cors_enabled" BOOLEAN NOT NULL DEFAULT TRUE,
  "routes" JSON,
  UNIQUE ("api_id", "name"),
  FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE,
  FOREIGN KEY("endpoint_group_id") REFERENCES "endpoint_groups"("id") ON DELETE SET NULL,
  FOREIGN KEY("environment_id") REFERENCES "environments"("id") ON DELETE RESTRICT
);
