CREATE TABLE IF NOT EXISTS `custom_function_files` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `created_at` DATETIME,
  `updated_at` DATETIME,
  `custom_function_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `body` TEXT NOT NULL,
  UNIQUE (`custom_function_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`custom_function_id`) REFERENCES `custom_functions`(`id`) ON DELETE CASCADE
);
CREATE INDEX idx_custom_function_files_custom_function_id ON custom_function_files(custom_function_id);
