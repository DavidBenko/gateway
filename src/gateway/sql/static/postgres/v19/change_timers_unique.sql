ALTER TABLE timers DROP CONSTRAINT timers_name_key;
ALTER TABLE timers ADD UNIQUE("api_id", "name");
