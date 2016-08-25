CREATE TABLE IF NOT EXISTS "plans" (
  "id" SERIAL PRIMARY KEY,
  "name" TEXT NOT NULL,
  "stripe_name" TEXT NOT NULL,
  "max_users" INTEGER NOT NULL DEFAULT 1,
  "javascript_timeout" INTEGER NOT NULL DEFAULT 5,
  "price" INTEGER NOT NULL DEFAULT 0,
  UNIQUE ("name")
);
CREATE INDEX idx_plan_name ON plans USING btree(name);
CREATE INDEX idx_plan_stripe_name ON plans USING btree(stripe_name);
