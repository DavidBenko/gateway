CREATE TABLE IF NOT EXISTS `push_messages` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `push_device_id` INTEGER NOT NULL,
  `stamp` INTEGER NOT NULL,
  `data` TEXT NOT NULL,
  FOREIGN KEY(`push_device_id`) REFERENCES `push_devices`(`id`) ON DELETE CASCADE
);
CREATE INDEX idx_push_messages_push_device_id ON push_messages(push_device_id);
