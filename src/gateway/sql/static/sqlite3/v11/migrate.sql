CREATE TABLE IF NOT EXISTS `_proxy_endpoint_components` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `endpoint_id` INTEGER NOT NULL,
  `conditional` TEXT,
  `conditional_positive` BOOLEAN DEFAULT 1,
  `position` INTEGER NOT NULL,
  `type` TEXT NOT NULL,
  `data` TEXT,
  `type_discriminator` TEXT,
  `name` TEXT,
  `description` TEXT,
  `api_id` INTEGER
);

INSERT INTO _proxy_endpoint_components(id, endpoint_id, conditional, conditional_positive, position, type, data)
SELECT id, endpoint_id, conditional, conditional_positive, position, type, data FROM proxy_endpoint_components;

DROP TABLE proxy_endpoint_components;

UPDATE _proxy_endpoint_components
SET api_id = (SELECT pe.api_id FROM proxy_endpoints pe WHERE pe.id = endpoint_id), type_discriminator = 'standard';

CREATE TABLE `proxy_endpoint_components` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `conditional` TEXT,
  `conditional_positive` BOOLEAN DEFAULT 1,
  `type` TEXT NOT NULL,
  `data` TEXT,
  `type_discriminator` TEXT,
  `name` TEXT,
  `description` TEXT,
  `api_id` INTEGER,
  UNIQUE (`api_id`, `name`) ON CONFLICT FAIL,
  FOREIGN KEY(`api_id`) REFERENCES `apis`(`id`) ON DELETE CASCADE
);

INSERT INTO proxy_endpoint_components(id, conditional, conditional_positive, type, data, type_discriminator, name, description, api_id)
SELECT id, conditional, conditional_positive, type, data, type_discriminator, name, description, api_id
FROM _proxy_endpoint_components;

CREATE TABLE IF NOT EXISTS `proxy_endpoint_component_references` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `proxy_endpoint_id` INTEGER NOT NULL,
  `position` INTEGER NOT NULL,
  `proxy_endpoint_component_id` INTEGER NOT NULL,
  FOREIGN KEY(`proxy_endpoint_id`) REFERENCES `proxy_endpoints`(`id`) ON DELETE CASCADE,
  FOREIGN KEY(`proxy_endpoint_component_id`) REFERENCES `proxy_endpoint_components`(`id`) ON DELETE NO ACTION
);

INSERT INTO proxy_endpoint_component_references(proxy_endpoint_id, position, proxy_endpoint_component_id)
SELECT endpoint_id, position, id as proxy_endpoint_component_id
FROM _proxy_endpoint_components;

CREATE TRIGGER proxy_endpoint_component_references_conditional_delete_trigger
AFTER DELETE ON proxy_endpoint_component_references
FOR EACH ROW
BEGIN
  DELETE FROM proxy_endpoint_components
  WHERE type_discriminator = 'standard'
  AND id IN (
    SELECT t.id FROM (
      SELECT OLD.proxy_endpoint_component_id as id, COUNT(1) as cnt
      FROM proxy_endpoint_component_references pecr
      WHERE pecr.proxy_endpoint_component_id = OLD.proxy_endpoint_component_id
    ) t
    WHERE t.cnt = 0
  );
END;

DROP TABLE _proxy_endpoint_components;

CREATE INDEX idx_proxy_endpoint_components_type_discriminator ON proxy_endpoint_components(type_discriminator);
CREATE INDEX idx_proxy_endpoint_components_api_id ON proxy_endpoint_components(api_id);
CREATE INDEX idx_proxy_endpoint_component_references_proxy_endpoint_id ON proxy_endpoint_component_references(proxy_endpoint_id);
CREATE INDEX idx_proxy_endpoint_component_references_proxy_endpoint_component_id ON proxy_endpoint_component_references(proxy_endpoint_component_id);

ANALYZE;
