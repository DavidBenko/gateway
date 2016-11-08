
ALTER TABLE `soap_remote_endpoints`  RENAME TO `tmp_soap_remote_endpoints`;

CREATE TABLE IF NOT EXISTS `soap_remote_endpoints` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `remote_endpoint_id` INTEGER UNIQUE NOT NULL,
  `wsdl` TEXT NOT NULL,
  `wsdl_content` BLOB,
  `wsdl_content_thumbprint` TEXT,
  FOREIGN KEY(`remote_endpoint_id`) REFERENCES `remote_endpoints`(`id`) ON DELETE CASCADE
);

INSERT INTO `soap_remote_endpoints`(`id`, `remote_endpoint_id`, `wsdl`, `wsdl_content`, `wsdl_content_thumbprint`)
SELECT `id`, `remote_endpoint_id`, `wsdl`, `generated_jar`, `generated_jar_thumbprint`
FROM `tmp_soap_remote_endpoints`;

DROP TABLE `tmp_soap_remote_endpoints`;



