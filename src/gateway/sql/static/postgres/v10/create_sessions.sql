CREATE TABLE IF NOT EXISTS "sessions" (
  "id" SERIAL PRIMARY KEY,
  "session_name" TEXT NOT NULL,
  "session_uuid" TEXT NOT NULL,
  "max_age" INTEGER NOT NULL,
  "expires" INTEGER NOT NULL,
  "session_values" JSONB NOT NULL,
  UNIQUE ("session_name", "session_uuid")
);

CREATE INDEX idx_sessions_session_name ON sessions USING btree(session_name);
CREATE INDEX idx_sessions_session_uuid ON sessions USING btree(session_uuid);
