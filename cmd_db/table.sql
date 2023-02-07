 

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint(20)  NOT NULL AUTO_INCREMENT,
  `email` varchar(200) DEFAULT NULL,
  `password` varchar(200) DEFAULT NULL,
  `token` varchar(32) DEFAULT NULL,
  `forbidden` tinyint(1) DEFAULT NULL,
  `roles` longtext,
  `permissions` longtext,
  `updated_unixtime` bigint(20) DEFAULT NULL,
  `created_unixtime` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `token` (`token`),
  KEY `idx_users_email` (`email`),
  KEY `idx_users_token` (`token`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4;


DROP TABLE IF EXISTS `dbkv`;
CREATE TABLE `dbkv` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `key` varchar(191) DEFAULT NULL,
  `value` longtext,
  `description` longtext,
  PRIMARY KEY (`id`),
  UNIQUE KEY `key` (`key`),
  KEY `idx_dbkv_key` (`key`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4;