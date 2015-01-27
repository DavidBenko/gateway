CREATE TABLE IF NOT EXISTS "endpoint_groups" (
  "id" SERIAL PRIMARY KEY,
  "api_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "description" TEXT,
  FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE
);
