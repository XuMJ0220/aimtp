-- Active: 1771131057693@@192.168.124.6@3306@aimtp
-- DAG 状态摘要表
-- dag_status_summary 表

CREATE TABLE IF NOT EXISTS `dag_status_summary` (
    `dag_id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'DAG自增ID',
    `dag_name` VARCHAR(255) UNIQUE NOT NULL COMMENT 'DAG名称(唯一)',
    `cluster` VARCHAR(64) NOT NULL COMMENT '所属集群',
    `user_name` VARCHAR(128) NOT NULL COMMENT 'DAG所属用户',
    `queue_name` VARCHAR(128) COMMENT 'DAG所属队列',
    `engine` VARCHAR(32) DEFAULT 'volcano' COMMENT '执行引擎: volcano/argo',

    `state` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'DAG状态: pending/creating/running/succeeded/failed',

    `creation_status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '创建状态: pending/creating/created/failed',
    `payload` MEDIUMTEXT COMMENT 'DAG定义JSON (创建后可清空)',
    `error_msg` TEXT COMMENT '创建失败原因',
    `retry_count` TINYINT DEFAULT 0 COMMENT '重新创建次数',
    `max_retries` TINYINT DEFAULT 3 COMMENT '最大重新创建次数',

    `total_jobs` INT DEFAULT 0 COMMENT '总Job数量',
    `completed_jobs` INT DEFAULT 0 COMMENT '已完成Job数量',
    `failed_jobs` INT DEFAULT 0 COMMENT '失败Job数量',

    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `started_at` TIMESTAMP NULL COMMENT 'DAG开始运行时间',
    `finished_at` TIMESTAMP NULL COMMENT 'DAG结束时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP NULL COMMENT '删除时间',
    -- 版本控制
    `resource_version` VARCHAR(64) COMMENT 'Kubernetes ResourceVersion',
    `version` BIGINT DEFAULT 0 COMMENT '版本号（纳秒时间戳）',

    INDEX `idx_cluster` (`cluster`),
    INDEX `idx_user` (`user_name`),
    INDEX `idx_state` (`state`),
    INDEX `idx_creation_queue` (`creation_status`, `created_at`),
    INDEX `idx_user_state` (`user_name`, `state`),
    INDEX `idx_updated` (`updated_at`),
    INDEX `idx_deleted` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='DAG状态摘要表';