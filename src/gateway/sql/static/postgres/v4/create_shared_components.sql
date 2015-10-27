CREATE TABLE IF NOT EXISTS "shared_components" (
  -- ProxyEndpointComponent
  "id" SERIAL PRIMARY KEY,
  "conditional" TEXT,
  "conditional_positive" BOOLEAN DEFAULT TRUE,
  "type" TEXT NOT NULL,
  "data" JSONB,
  -- SharedComponent
  "api_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "description" TEXT,
  UNIQUE ("api_id", "name"),
  -- FK Constraint
  FOREIGN KEY("api_id") 
    REFERENCES "apis"("id") 
    ON DELETE CASCADE
);
