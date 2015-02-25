CREATE TABLE IF NOT EXISTS "proxy_endpoint_transformations" (
  "id" SERIAL PRIMARY KEY,
  "component_id" INTEGER,
  "call_id" INTEGER,
  "before" BOOLEAN NOT NULL DEFAULT TRUE,
  "position" INTEGER NOT NULL,
  "type" TEXT NOT NULL,
  "data" JSONB,
  FOREIGN KEY("component_id") REFERENCES "proxy_endpoint_components"("id") ON DELETE CASCADE,
  FOREIGN KEY("call_id") REFERENCES "proxy_endpoint_calls"("id") ON DELETE CASCADE
);
