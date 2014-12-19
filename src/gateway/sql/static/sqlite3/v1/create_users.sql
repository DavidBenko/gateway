CREATE TABLE IF NOT EXISTS `users` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `account_id` INTEGER NOT NULL,
  `name` TEXT NOT NULL,
  `email` TEXT UNIQUE NOT NULL,
  `password` TEXT NOT NULL,
  FOREIGN KEY(`account_id`) REFERENCES `accounts`(`id`)
);
