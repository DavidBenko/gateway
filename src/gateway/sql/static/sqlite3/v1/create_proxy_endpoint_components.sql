CREATE TABLE IF NOT EXISTS `proxy_endpoint_components` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `endpoint_id` INTEGER NOT NULL,
  `conditional` TEXT,
  `conditional_positive` BOOLEAN DEFAULT 1,
  `position` INTEGER NOT NULL,
  `type` TEXT NOT NULL,
  `data` TEXT,
  FOREIGN KEY(`endpoint_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE
);
