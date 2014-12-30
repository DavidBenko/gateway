CREATE TABLE IF NOT EXISTS `environment_values` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `environment_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `value` TEXT,
  FOREIGN KEY(`environment_id`) REFERENCES `environments`(`id`) ON DELETE CASCADE
);
