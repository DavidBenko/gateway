CREATE TABLE IF NOT EXISTS `push_devices` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `push_channel_id` INTEGER NOT NULL,
  `environment_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `type` TEXT NOT NULL,
  `token` TEXT NOT NULL,
  `expires` INTEGER NOT NULL,
  `data` TEXT NOT NULL,
  UNIQUE (`push_channel_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`environment_id`) REFERENCES `environments`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`push_channel_id`) REFERENCES `push_channels`(`id`) ON DELETE CASCADE
);
CREATE INDEX idx_push_devices_token ON push_devices(token);
