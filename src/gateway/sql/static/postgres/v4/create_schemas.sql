CREATE TABLE IF NOT EXISTS "schemas" (
  "id" SERIAL PRIMARY KEY,
  "api_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "description" TEXT,
  "type" TEXT NOT NULL,
  "schema" TEXT,
  "data" TEXT,
  UNIQUE ("api_id", "name"),
  FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE
);
