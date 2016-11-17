CREATE TABLE IF NOT EXISTS `proxy_endpoint_channels` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `created_at` DATETIME,
  `updated_at` DATETIME,
  `proxy_endpoint_id` INTEGER NOT NULL,
  `remote_endpoint_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  UNIQUE (`remote_endpoint_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`proxy_endpoint_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`remote_endpoint_id`) REFERENCES `remote_endpoints`(`id`) ON DELETE CASCADE
);
