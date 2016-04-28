CREATE TABLE IF NOT EXISTS `push_channels` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `account_id` INTEGER NOT NULL,
  `api_id` INTEGER NOT NULL,
  `remote_endpoint_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `expires` INTEGER NOT NULL,
  `data` TEXT NOT NULL,
  UNIQUE (`remote_endpoint_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`account_id`) REFERENCES `accounts`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`remote_endpoint_id`) REFERENCES `remote_endpoints`(`id`) ON DELETE CASCADE
);
