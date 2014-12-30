CREATE TABLE IF NOT EXISTS `routes` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `api_id` INTEGER NOT NULL,
  `endpoint_id` INTEGER NOT NULL,
  `method` TEXT NOT NULL,
  `path` TEXT NOT NULL,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`endpoint_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE
);
