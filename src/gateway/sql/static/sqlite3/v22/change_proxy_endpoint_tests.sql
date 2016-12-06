CREATE TABLE IF NOT EXISTS `_proxy_endpoint_tests` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `created_at` DATETIME,
  `updated_at` DATETIME,
  `endpoint_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `channels` BOOLEAN NOT NULL,
  `channel_id` INTEGER,
  `methods` TEXT NOT NULL,
  `route` TEXT NOT NULL,
  `body` TEXT,
  `data` TEXT,
  UNIQUE (`endpoint_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`channel_id`) REFERENCES `proxy_endpoint_channels`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`endpoint_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE
);

INSERT INTO _proxy_endpoint_tests(id, endpoint_id, name, channels, methods, route, body, data)
SELECT id, endpoint_id, name, 0 as channels, methods, route, body, data FROM proxy_endpoint_tests;

DROP TABLE proxy_endpoint_tests;
ALTER TABLE _proxy_endpoint_tests RENAME TO proxy_endpoint_tests;

CREATE INDEX idx_proxy_endpoint_tests_endpoint_id ON proxy_endpoint_tests(endpoint_id);
