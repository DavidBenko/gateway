CREATE TABLE IF NOT EXISTS "environments" (
  "id" SERIAL PRIMARY KEY,
  "api_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "description" TEXT,
  "data" TEXT NOT NULL,
  UNIQUE ("api_id", "name"),
  FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE
);
