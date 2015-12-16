-- Temporary table to house accounts data
CREATE TABLE `_accounts` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `name` TEXT NOT NULL
);

-- Insert all account data into the temp accounts table
INSERT INTO _accounts(id, name) SELECT id, name FROM accounts;

-- Drop old accounts table so we can create a new one without UNIQUE constraint on 'name'
DROP TABLE accounts;

-- Recreate new accounts table without UNIQUE constraint on 'name'
CREATE TABLE `accounts` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `name` TEXT NOT NULL
);

-- Copy data from temp accounts table into new accounts table
INSERT INTO accounts(id, name) SELECT id, name FROM _accounts;

-- Drop temp table
DROP TABLE _accounts;
