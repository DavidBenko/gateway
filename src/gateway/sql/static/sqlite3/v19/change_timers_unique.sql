CREATE TABLE IF NOT EXISTS `_timers` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `created_at` DATETIME,
  `updated_at` DATETIME,
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
  `parameters` TEXT NOT NULL,
  `data` TEXT NOT NULL,
  UNIQUE (`api_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`job_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE
);

INSERT INTO _timers(id, created_at, updated_at, api_id, job_id, name, once, time_zone, minute, hour, day_of_month, month, day_of_week, next, parameters, data)
SELECT id, created_at, updated_at, api_id, job_id, name, once, time_zone, minute, hour, day_of_month, month, day_of_week, next, parameters, data FROM timers;

DROP TABLE timers;
ALTER TABLE _timers RENAME TO timers;
