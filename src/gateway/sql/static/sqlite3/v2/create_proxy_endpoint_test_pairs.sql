CREATE TABLE IF NOT EXISTS `proxy_endpoint_test_pairs` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `test_id` INTEGER NOT NULL,
  `type` TEXT NOT NULL,
  `key` TEXT NOT NULL,
  `value` TEXT,
  FOREIGN KEY(`test_id`) REFERENCES `proxy_endpoint_tests`(`id`) ON DELETE CASCADE
);
