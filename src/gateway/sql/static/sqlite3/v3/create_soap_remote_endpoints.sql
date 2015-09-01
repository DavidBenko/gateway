CREATE TABLE IF NOT EXISTS `soap_remote_endpoints` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `remote_endpoint_id` INTEGER NOT NULL,
  `wsdl` TEXT NOT NULL,
  `generated_jar` BLOB,
  `generated_jar_thumbprint` TEXT,
  FOREIGN KEY(`remote_endpoint_id`) REFERENCES `remote_endpoints`(`id`) ON DELETE CASCADE
);

ALTER TABLE remote_endpoints ADD COLUMN status TEXT;
ALTER TABLE remote_endpoints ADD COLUMN status_message TEXT;
