CREATE TABLE IF NOT EXISTS "proxy_endpoint_channels" (
  "id" SERIAL PRIMARY KEY,
  "proxy_endpoint_id" INTEGER NOT NULL,
  "remote_endpoint_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  UNIQUE ("remote_endpoint_id", "name"),
  FOREIGN KEY("proxy_endpoint_id") REFERENCES "proxy_endpoints"("id") ON DELETE CASCADE,
  FOREIGN KEY("remote_endpoint_id") REFERENCES "remote_endpoints"("id") ON DELETE CASCADE
);
