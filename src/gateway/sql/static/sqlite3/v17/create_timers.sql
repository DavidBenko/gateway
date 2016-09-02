CREATE TABLE IF NOT EXISTS `timers` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `api_id` INTEGER NOT NULL,
  `job_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `once` BOOLEAN NOT NULL DEFAULT 0,
  `time_zone` INTEGER NOT NULL,
  `minute` TEXT NOT NULL,
  `hour` TEXT NOT NULL,
  `day_of_month` TEXT NOT NULL,
  `month` TEXT NOT NULL,
  `day_of_week` TEXT NOT NULL,
  `next` INTEGER NOT NULL,
  `attributes` TEXT NOT NULL,
  `data` TEXT NOT NULL,
  UNIQUE (`name`) ON CONFLICT FAIL,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`job_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE
);
CREATE INDEX idx_timers_api_id ON timers(api_id);
CREATE INDEX idx_timers_next ON timers(next);
