CREATE TABLE IF NOT EXISTS `job_tests` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `created_at` DATETIME,
  `updated_at` DATETIME,
  `job_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `parameters` TEXT NOT NULL,
  UNIQUE (`job_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`job_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE
);
