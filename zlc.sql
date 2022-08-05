/*
 Navicat Premium Data Transfer

 Source Server         : 137.184.120.146
 Source Server Type    : MySQL
 Source Server Version : 50734
 Source Host           : 137.184.120.146:3306
 Source Schema         : zlc

 Target Server Type    : MySQL
 Target Server Version : 50734
 File Encoding         : 65001

 Date: 08/07/2022 16:35:59
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `id` int(8) NOT NULL AUTO_INCREMENT,
  `tgid` bigint(12) NULL DEFAULT NULL,
  `binded` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `isGroup` int(1) NULL DEFAULT 0 COMMENT '是否进群',
  `bean` int(1) NULL DEFAULT 0,
  `ddfactory` int(1) NULL DEFAULT 0,
  `farm` int(1) NULL DEFAULT 0,
  `health` int(1) NULL DEFAULT 0,
  `jxfactory` int(1) NULL DEFAULT 0,
  `pet` int(1) NULL DEFAULT 0,
  `sgmh` int(1) NULL DEFAULT 0,
  `cfd` int(1) NULL DEFAULT 0,
  `city` int(1) NULL DEFAULT 0,
  `carnivalcity` int(1) NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 12989 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
