CREATE TABLE IF NOT EXISTS `environments` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `api_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `description` TEXT,
  `data` TEXT NOT NULL,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE
);
