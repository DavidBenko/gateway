CREATE TABLE IF NOT EXISTS "custom_functions" (
  "id" SERIAL PRIMARY KEY,
  "created_at" TIMESTAMPTZ,
  "updated_at" TIMESTAMPTZ,
  "api_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "description" TEXT,
  "active" BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE ("api_id", "name"),
  FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE
);
CREATE INDEX idx_custom_functions_api_id ON custom_functions USING btree(api_id);
