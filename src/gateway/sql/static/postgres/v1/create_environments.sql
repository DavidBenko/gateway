CREATE TABLE IF NOT EXISTS "environments" (
  "id" SERIAL PRIMARY KEY,
  "api_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "description" TEXT,
  "data" JSON NOT NULL,
  "session_auth_key" TEXT,
  "session_encryption_key" TEXT,
  "session_auth_key_rotate" TEXT,
  "session_encryption_key_rotate" TEXT,
  UNIQUE ("api_id", "name"),
  FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE
);
