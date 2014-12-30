CREATE TABLE IF NOT EXISTS `proxy_endpoint_components` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `endpoint_id` INTEGER NOT NULL,
  `conditional` TEXT,
  `conditional_negate` BOOLEAN DEFAULT 0,
  `position` INTEGER NOT NULL,
  `type` TEXT NOT NULL,
  FOREIGN KEY(`endpoint_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE
);
