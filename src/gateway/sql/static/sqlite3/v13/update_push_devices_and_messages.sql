DROP TABLE push_messages;

DROP TABLE push_devices;

CREATE TABLE IF NOT EXISTS `push_devices` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `remote_endpoint_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `type` TEXT NOT NULL,
  `token` TEXT NOT NULL,
  `data` TEXT NOT NULL,
  UNIQUE (`remote_endpoint_id`, `type`, 'token') ON CONFLICT FAIL,
  FOREIGN KEY(`remote_endpoint_id`) REFERENCES `remote_endpoints`(`id`) ON DELETE CASCADE
);
CREATE INDEX idx_push_devices_token ON push_devices(token);

CREATE TABLE IF NOT EXISTS `push_channels_push_devices` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `push_channel_id` INTEGER NOT NULL,
  `push_device_id` INTEGER NOT NULL,
  `expires` INTEGER NOT NULL,
  UNIQUE (`push_channel_id`, `push_device_id`) ON CONFLICT FAIL,
  FOREIGN KEY(`push_channel_id`) REFERENCES `push_channels`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`push_device_id`) REFERENCES `push_devices`(`id`) ON DELETE CASCADE
);
CREATE INDEX idx_push_channels_push_devices ON push_channels_push_devices(push_channel_id, push_device_id);

CREATE TABLE IF NOT EXISTS `push_messages` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `push_device_id` INTEGER NOT NULL,
  `push_channel_id` INTEGER,
  `push_channel_message_id` INTEGER,
  `stamp` INTEGER NOT NULL,
  `data` TEXT NOT NULL,
  FOREIGN KEY(`push_device_id`) REFERENCES `push_devices`(`id`) ON DELETE CASCADE
  FOREIGN KEY(`push_channel_id`) REFERENCES `push_channels`(`id`) ON DELETE SET NULL
  FOREIGN KEY(`push_channel_message_id`) REFERENCES `push_channel_messages`(`id`) ON DELETE SET NULL
);
CREATE INDEX idx_push_messages_push_device_id ON push_messages(push_device_id);
CREATE INDEX idx_push_messages_push_channel_id ON push_messages(push_channel_id);
CREATE INDEX idx_push_messages_push_channel_message_id ON push_messages(push_channel_message_id);

CREATE TABLE IF NOT EXISTS `push_channel_messages` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `push_channel_id` INTEGER NOT NULL,
  `stamp` INTEGER NOT NULL,
  `data` TEXT NOT NULL,
  FOREIGN KEY(`push_channel_id`) REFERENCES `push_channels`(`id`) ON DELETE CASCADE
);
CREATE INDEX idx_push_channel_messages_push_channel_id ON push_messages(push_channel_id);
