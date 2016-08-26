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
  `name` TEXT NOT NULL,
  `plan_id` INTEGER,
  `stripe_customer_id` TEXT,
  `stripe_subscription_id` TEXT,
  `stripe_payment_retry_attempt` INTEGER DEFAULT 0,
  UNIQUE (`stripe_customer_id`) ON CONFLICT FAIL,
  UNIQUE (`stripe_subscription_id`) ON CONFLICT FAIL,
  FOREIGN KEY(`plan_id`) REFERENCES `plans`(`id`) ON DELETE SET NULL
);
CREATE INDEX idx_account_stripe_customer_id ON accounts(stripe_customer_id);
CREATE INDEX idx_account_stripe_subscription_id ON accounts(stripe_subscription_id);

-- Copy data from temp accounts table into new accounts table
INSERT INTO accounts(id, name) SELECT id, name FROM _accounts;

-- Drop temp table
DROP TABLE _accounts;
