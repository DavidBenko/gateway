CREATE TABLE IF NOT EXISTS `plans` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `name` TEXT NOT NULL,
  `stripe_name` TEXT NOT NULL,
  `max_users` INTEGER NOT NULL DEFAULT 1,
  `javascript_timeout` INTEGER NOT NULL DEFAULT 5,
  `price` INTEGER NOT NULL DEFAULT 0,
  UNIQUE (`name`) ON CONFLICT FAIL
);
CREATE INDEX idx_plan_name ON plans(name);
CREATE INDEX idx_plan_stripe_name ON plans(stripe_name);
