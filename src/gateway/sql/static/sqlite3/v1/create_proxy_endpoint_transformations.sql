CREATE TABLE IF NOT EXISTS `proxy_endpoint_transformations` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `component_id` INTEGER,
  `call_id` INTEGER,
  `before` BOOLEAN NOT NULL DEFAULT 1,
  `position` INTEGER NOT NULL,
  `type` TEXT NOT NULL,
  `data` TEXT,
  FOREIGN KEY(`component_id`) REFERENCES `proxy_endpoint_components`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`call_id`) REFERENCES `proxy_endpoint_calls`(`id`) ON DELETE CASCADE
);
