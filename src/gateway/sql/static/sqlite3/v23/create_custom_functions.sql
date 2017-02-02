CREATE TABLE IF NOT EXISTS `custom_functions` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `created_at` DATETIME,
  `updated_at` DATETIME,
  `api_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `description` TEXT,
  `active` BOOLEAN NOT NULL DEFAULT 1,
  'memory' INTEGER NOT NULL,
  'cpu_shares' INTEGER NOT NULL,
  'timeout' INTEGER NOT NULL,
  UNIQUE (`api_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE
);
CREATE INDEX idx_custom_functions_api_id ON custom_functions(api_id);
