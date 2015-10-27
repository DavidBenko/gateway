CREATE TABLE IF NOT EXISTS `shared_components` (
  -- ProxyEndpointComponent
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `conditional` TEXT,
  `conditional_positive` BOOLEAN DEFAULT 1,
  `type` TEXT NOT NULL,
  `data` TEXT,
  -- SharedComponent
  `api_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `description` TEXT,
  UNIQUE (`api_id`, `name`)
    ON CONFLICT FAIL,
  -- FK Constraint
  FOREIGN KEY(`api_id`)
    REFERENCES `apis`(`id`)
    ON DELETE CASCADE
);
