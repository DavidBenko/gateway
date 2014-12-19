CREATE TABLE IF NOT EXISTS `users` (
  `id` INTEGER SERIAL PRIMARY KEY ,
  `account_id` INTEGER NOT NULL REFERENCES `accounts`(`id`),
  `name` TEXT NOT NULL,
  `email` TEXT NOT NULL UNIQUE,
  `password` TEXT NOT NULL
);
