CREATE TABLE IF NOT EXISTS `proxy_endpoint_tests` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `endpoint_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `methods` TEXT NOT NULL,
  `route` TEXT NOT NULL,
  `body` TEXT,
  `data` TEXT,
  UNIQUE (`endpoint_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`endpoint_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE
);
