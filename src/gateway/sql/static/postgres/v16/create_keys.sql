CREATE TABLE IF NOT EXISTS "keys" (
  "id" SERIAL PRIMARY KEY,
  "account_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "key" bytea NOT NULL,
  UNIQUE ("account_id", "name"),
  FOREIGN KEY("account_id") REFERENCES "accounts"("id") ON DELETE CASCADE
);
