CREATE TABLE `stats` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT
  , `node` TEXT NOT NULL
  , `ms` BIGINT NOT NULL
  , `api_id` INT NOT NULL
  , `request_size` INT NOT NULL
  , `request_id` TEXT NOT NULL
  , `response_time` INT NOT NULL
  , `response_size` INT NOT NULL
  , `response_status` INT NOT NULL
  , `response_error` TEXT NOT NULL
  , `timestamp` TIMESTAMP NOT NULL
)
