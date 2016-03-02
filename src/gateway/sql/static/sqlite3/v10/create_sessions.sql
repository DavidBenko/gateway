CREATE TABLE IF NOT EXISTS `sessions` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `session_name` TEXT NOT NULL,
  `session_uuid` TEXT NOT NULL,
  `max_age` INTEGER NOT NULL,
  `expires` INTEGER NOT NULL,
  `session_values` TEXT NOT NULL,
  UNIQUE (`session_name`, `session_uuid`) ON CONFLICT FAIL
);

CREATE INDEX idx_sessions_session_name ON sessions(session_name);
CREATE INDEX idx_sessions_session_uuid ON sessions(session_uuid);
