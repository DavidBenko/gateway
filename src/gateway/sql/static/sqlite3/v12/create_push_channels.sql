CREATE TABLE IF NOT EXISTS `push_channels` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `remote_endpoint_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `expires` INTEGER NOT NULL,
  `data` TEXT NOT NULL,
  UNIQUE (`remote_endpoint_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`remote_endpoint_id`) REFERENCES `remote_endpoints`(`id`) ON DELETE CASCADE
);
