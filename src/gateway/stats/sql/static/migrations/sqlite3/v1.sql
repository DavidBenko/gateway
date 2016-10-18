CREATE TABLE `stats` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT
  , `node` TEXT NOT NULL
  , `ms` BIGINT NOT NULL
  , `request_size` INT NOT NULL
  , `request_id` TEXT NOT NULL
  , `api_id` INT NOT NULL
  , `api_name` TEXT NOT NULL
  , `host_id` INT NOT NULL
  , `host_name` TEXT NOT NULL
  , `proxy_id` INT NOT NULL
  , `proxy_name` TEXT NOT NULL
  , `proxy_env_id` INT NOT NULL
  , `proxy_env_name` TEXT NOT NULL
  , `proxy_route_path` TEXT NOT NULL
  , `proxy_route_verb` TEXT NOT NULL
  , `proxy_group_id` INT
  , `proxy_group_name` TEXT
  , `response_time` INT NOT NULL
  , `response_size` INT NOT NULL
  , `response_status` INT NOT NULL
  , `response_error` TEXT NOT NULL
  , `remote_endpoint_response_time` INT NOT NULL
  , `timestamp` TIMESTAMP NOT NULL
);

CREATE INDEX idx_stats_api_id ON stats(api_id);
CREATE INDEX idx_stats_proxy_id ON stats(proxy_id);
CREATE INDEX idx_stats_timestamp ON stats(timestamp);
