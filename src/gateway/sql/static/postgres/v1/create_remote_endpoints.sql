CREATE TABLE IF NOT EXISTS "remote_endpoints" (
  "id" SERIAL PRIMARY KEY,
  "api_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "description" TEXT,
  "type" TEXT NOT NULL,
  FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE
);
