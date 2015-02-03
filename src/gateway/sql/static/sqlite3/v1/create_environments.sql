CREATE TABLE IF NOT EXISTS `environments` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `api_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `description` TEXT,
  `data` TEXT NOT NULL,
  UNIQUE (`api_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE
);
