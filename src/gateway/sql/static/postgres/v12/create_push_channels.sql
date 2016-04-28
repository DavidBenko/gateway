CREATE TABLE IF NOT EXISTS "push_channels" (
  "id" SERIAL PRIMARY KEY,
  "account_id" INTEGER NOT NULL,
  "api_id" INTEGER NOT NULL,
  "remote_endpoint_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "expires" INTEGER NOT NULL,
  "data" JSONB NOT NULL,
  UNIQUE ("remote_endpoint_id", "name"),
  FOREIGN KEY("account_id") REFERENCES "accounts"("id") ON DELETE CASCADE,
  FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE,
  FOREIGN KEY("remote_endpoint_id") REFERENCES "remote_endpoints"("id") ON DELETE CASCADE
);
