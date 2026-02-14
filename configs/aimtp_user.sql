-- AIMTP User Schema
-- MySQL dump for aimtp project
-- Table: user

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Create Database
-- ----------------------------
CREATE DATABASE IF NOT EXISTS `aimtp` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `aimtp`;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `userID` varchar(36) NOT NULL DEFAULT '' COMMENT '用户唯一 ID',
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名（唯一）',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码（加密后）',
  `nickname` varchar(30) NOT NULL DEFAULT '' COMMENT '用户昵称',
  `email` varchar(256) NOT NULL DEFAULT '' COMMENT '用户电子邮箱地址',
  `phone` varchar(16) NOT NULL DEFAULT '' COMMENT '用户手机号',
  `createdAt` datetime NOT NULL DEFAULT current_timestamp() COMMENT '用户创建时间',
  `updatedAt` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '用户最后修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `user.userID` (`userID`),
  UNIQUE KEY `user.username` (`username`),
  UNIQUE KEY `user.phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

SET FOREIGN_KEY_CHECKS = 1;