CREATE TABLE IF NOT EXISTS "apis" (
  "id" SERIAL PRIMARY KEY,
  "account_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "description" TEXT,
  "cors_allow" TEXT DEFAULT '*',
  FOREIGN KEY("account_id") REFERENCES "accounts"("id") ON DELETE CASCADE
);
