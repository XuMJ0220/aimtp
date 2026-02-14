-- AIMTP Casbin Schema
-- MySQL dump for aimtp project
-- Table: casbin_rule

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Create Database
-- ----------------------------
CREATE DATABASE IF NOT EXISTS `aimtp` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `aimtp`;

-- ----------------------------
-- Table structure for casbin_rule
-- ----------------------------
DROP TABLE IF EXISTS `casbin_rule`;
CREATE TABLE `casbin_rule` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `ptype` varchar(100) DEFAULT NULL,
  `v0` varchar(100) DEFAULT NULL,
  `v1` varchar(100) DEFAULT NULL,
  `v2` varchar(100) DEFAULT NULL,
  `v3` varchar(100) DEFAULT NULL,
  `v4` varchar(100) DEFAULT NULL,
  `v5` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- Records of casbin_rule
-- ----------------------------
INSERT INTO `casbin_rule` VALUES
(18,'g','user-000000','role::admin',NULL,NULL,'',''),
(21,'p','role::admin','*','*','allow','',''),
(7,'p','role::user','/v1.AIMTP/DeleteUser','CALL','deny','',''),
(8,'p','role::user','/v1.AIMTP/ListUser','CALL','deny','',''),
(9,'p','role::user','/v1/users','GET','deny','',''),
(10,'p','role::user','/v1/users/*','DELETE','deny','','');

SET FOREIGN_KEY_CHECKS = 1;