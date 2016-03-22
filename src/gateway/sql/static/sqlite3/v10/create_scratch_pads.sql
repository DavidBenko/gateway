CREATE TABLE IF NOT EXISTS `scratch_pads` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `environment_data_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `code` TEXT NOT NULL,
  `data` TEXT NOT NULL,
  UNIQUE (`environment_data_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`environment_data_id`) REFERENCES `remote_endpoint_environment_data`(`id`) ON DELETE CASCADE
);
CREATE INDEX idx_scratch_pads_environment_data_id ON scratch_pads(environment_data_id);
