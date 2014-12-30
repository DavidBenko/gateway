CREATE TABLE IF NOT EXISTS `proxy_endpoint_calls` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `component_id` INTEGER NOT NULL,
  `conditional` TEXT,
  `conditional_negate` BOOLEAN DEFAULT 0,
  `position` INTEGER NOT NULL,
  FOREIGN KEY(`component_id`) REFERENCES `proxy_endpoint_components`(`id`) ON DELETE CASCADE
);
