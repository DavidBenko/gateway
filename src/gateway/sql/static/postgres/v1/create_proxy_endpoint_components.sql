CREATE TABLE IF NOT EXISTS "proxy_endpoint_components" (
  "id" SERIAL PRIMARY KEY,
  "endpoint_id" INTEGER NOT NULL,
  "conditional" TEXT,
  "conditional_positive" BOOLEAN DEFAULT TRUE,
  "position" INTEGER NOT NULL,
  "type" TEXT NOT NULL,
  "data" JSONB,
  FOREIGN KEY("endpoint_id") REFERENCES "proxy_endpoints"("id") ON DELETE CASCADE
);
