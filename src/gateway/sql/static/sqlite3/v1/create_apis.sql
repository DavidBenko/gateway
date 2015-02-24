CREATE TABLE IF NOT EXISTS `apis` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `account_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `description` TEXT,
  `cors_allow_origin` TEXT DEFAULT '*',
  `cors_allow_headers` TEXT DEFAULT 'content-type, accept',
  `cors_allow_credentials` BOOLEAN DEFAULT 1,
  `cors_request_headers` TEXT DEFAULT '*',
  `cors_max_age` INTEGER DEFAULT 600,
  UNIQUE (`account_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`account_id`) REFERENCES `accounts`(`id`) ON DELETE CASCADE
);
