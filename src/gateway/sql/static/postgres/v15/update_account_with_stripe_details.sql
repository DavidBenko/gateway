ALTER TABLE accounts ADD COLUMN plan_id INTEGER;
ALTER TABLE accounts ADD COLUMN stripe_customer_id TEXT;
ALTER TABLE accounts ADD COLUMN stripe_subscription_id TEXT;
ALTER TABLE accounts ADD COLUMN stripe_payment_retry_attempt INTEGER DEFAULT 0;
ALTER TABLE accounts ADD CONSTRAINT accounts_plan_fk FOREIGN KEY("plan_id") REFERENCES "plans"("id") ON DELETE SET NULL;
ALTER TABLE accounts ADD CONSTRAINT accounts_stripe_customer_unq UNIQUE ("stripe_customer_id");
ALTER TABLE accounts ADD CONSTRAINT accounts_stripe_sub_unq UNIQUE ("stripe_subscription_id");

CREATE INDEX idx_account_stripe_customer_id ON accounts USING btree(stripe_customer_id);
CREATE INDEX idx_account_stripe_subscription_id ON accounts USING btree(stripe_subscription_id);
