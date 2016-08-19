PRAGMA foreign_keys = OFF;
CREATE TABLE IF NOT EXISTS `_proxy_endpoints` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `api_id` INTEGER NOT NULL,
  `type` TEXT NOT NULL DEFAULT '',
  `endpoint_group_id` INTEGER,
  `environment_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `description` TEXT,
  `active` BOOLEAN NOT NULL DEFAULT 1,
  `cors_enabled` BOOLEAN NOT NULL DEFAULT 1,
  `routes` TEXT,
  UNIQUE (`api_id`, `type`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`endpoint_group_id`) REFERENCES `endpoint_groups`(`id`) ON DELETE SET NULL,
  FOREIGN KEY(`environment_id`) REFERENCES `environments`(`id`)
);

INSERT INTO _proxy_endpoints(id, api_id, type, endpoint_group_id, environment_id, name, description, active, cors_enabled, routes)
SELECT id, api_id, 'http' as type, endpoint_group_id, environment_id, name, description, active, cors_enabled, routes FROM proxy_endpoints;

DROP TABLE proxy_endpoints;
ALTER TABLE _proxy_endpoints RENAME TO proxy_endpoints;
PRAGMA foreign_keys = ON;
