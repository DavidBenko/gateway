CREATE TABLE IF NOT EXISTS "custom_function_tests" (
  "id" SERIAL PRIMARY KEY,
  "created_at" TIMESTAMPTZ,
  "updated_at" TIMESTAMPTZ,
  "custom_function_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "input" JSONB NOT NULL,
  UNIQUE ("custom_function_id", "name"),
  FOREIGN KEY("custom_function_id") REFERENCES "custom_functions"("id") ON DELETE CASCADE
);
