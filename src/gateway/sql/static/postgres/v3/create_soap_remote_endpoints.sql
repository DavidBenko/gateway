CREATE TABLE IF NOT EXISTS "soap_remote_endpoints" (
  "id" SERIAL PRIMARY KEY,
  "remote_endpoint_id" INTEGER NOT NULL,
  "wsdl" TEXT NOT NULL,
  "generated_jar" BYTEA,
  "status" TEXT NOT NULL DEFAULT 'Uninitialized',
  "message" TEXT,
  FOREIGN KEY("remote_endpoint_id") REFERENCES "remote_endpoints"("id") ON DELETE CASCADE
);
