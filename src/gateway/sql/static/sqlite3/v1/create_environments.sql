CREATE TABLE IF NOT EXISTS `environments` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `api_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `description` TEXT,
  `data` TEXT NOT NULL,
  `session_name` TEXT,
  `session_auth_key` TEXT,
  `session_encryption_key` TEXT,
  `session_auth_key_rotate` TEXT,
  `session_encryption_key_rotate` TEXT,
  UNIQUE (`api_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE
);
