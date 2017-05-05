CREATE TABLE IF NOT EXISTS "custom_function_files" (
  "id" SERIAL PRIMARY KEY,
  "created_at" TIMESTAMPTZ,
  "updated_at" TIMESTAMPTZ,
  "custom_function_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "body" TEXT NOT NULL,
  UNIQUE ("custom_function_id", "name"),
  FOREIGN KEY("custom_function_id") REFERENCES "custom_functions"("id") ON DELETE CASCADE
);
CREATE INDEX idx_custom_function_files_custom_function_id ON custom_function_files USING btree(custom_function_id);
