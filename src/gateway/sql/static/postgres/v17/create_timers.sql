CREATE TABLE IF NOT EXISTS "timers" (
  "id" SERIAL PRIMARY KEY,
  "api_id" INTEGER NOT NULL,
  "job_id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "once" BOOLEAN NOT NULL DEFAULT FALSE,
  "time_zone" INTEGER NOT NULL,
  "minute" TEXT NOT NULL,
  "hour" TEXT NOT NULL,
  "day_of_month" TEXT NOT NULL,
  "month" TEXT NOT NULL,
  "day_of_week" TEXT NOT NULL,
  "next" BIGINT NOT NULL,
  "parameters" JSONB NOT NULL,
  "data" JSONB NOT NULL,
  UNIQUE ("name"),
  FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE,
  FOREIGN KEY("job_id") REFERENCES "proxy_endpoints"("id") ON DELETE CASCADE
);
CREATE INDEX idx_timers_api_id ON timers USING btree(api_id);
CREATE INDEX idx_timers_next ON timers USING btree(next);
