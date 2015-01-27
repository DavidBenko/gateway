CREATE TABLE IF NOT EXISTS "proxy_endpoint_calls" (
  "id" SERIAL PRIMARY KEY,
  "component_id" INTEGER NOT NULL,
  "remote_endpoint_id" INTEGER NOT NULL,
  "endpoint_name_override" TEXT,
  "conditional" TEXT,
  "conditional_positive" BOOLEAN DEFAULT TRUE,
  "position" INTEGER NOT NULL,
  FOREIGN KEY("component_id") REFERENCES "proxy_endpoint_components"("id") ON DELETE CASCADE,
  FOREIGN KEY("remote_endpoint_id") REFERENCES "remote_endpoints"("id") ON DELETE RESTRICT
);
