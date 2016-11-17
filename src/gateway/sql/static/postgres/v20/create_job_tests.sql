CREATE TABLE IF NOT EXISTS "job_tests" (
  "id" SERIAL PRIMARY KEY,
  "created_at" TIMESTAMPTZ,
  "updated_at" TIMESTAMPTZ,
  "job_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "parameters" JSONB NOT NULL,
  UNIQUE ("job_id", "name"),
  FOREIGN KEY("job_id") REFERENCES "proxy_endpoints"("id") ON DELETE CASCADE
);
