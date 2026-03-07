-- Active: 1772191086950@@192.168.124.9@3306@aimtp
-- =============================================
-- Local: pod_status 表
-- 描述: Pod 状态表（一个 Job 可有多个 Pod）
-- =============================================

CREATE TABLE IF NOT EXISTS `pod_status` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `pod_name` VARCHAR(255) UNIQUE NOT NULL COMMENT 'Pod名称',
    -- 所属 Job 和 DAG
    `job_id` VARCHAR(255) NOT NULL COMMENT '所属Job ID',
    `dag_name` VARCHAR(255) NOT NULL COMMENT 'DAG名称',
    `cluster` VARCHAR(64) NOT NULL COMMENT '集群名称',
    -- Pod 标识
    `uuid` VARCHAR(64) COMMENT 'Job UUID（用于关联重试）',
    `replica_index` INT DEFAULT 0 COMMENT '副本索引（分布式训练）',
    `retry_index` INT DEFAULT 0 COMMENT '重试索引',
    `namespace` VARCHAR(128) DEFAULT 'default' COMMENT '命名空间',
    -- Pod 详细信息
    `node_name` VARCHAR(255) COMMENT '节点名称',
    `pod_ip` VARCHAR(64) COMMENT 'Pod IP',
    `host_ip` VARCHAR(64) COMMENT '主机IP',
    -- 状态信息
    `state` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'Pod状态：pending/running/succeeded/failed',
    `reason` VARCHAR(512) COMMENT '原因',
    `message` TEXT COMMENT '消息',
    `exit_code` INT COMMENT '退出码',
    -- 资源与设备
    `device` VARCHAR(512) COMMENT '分配的设备（GPU等）',
    `resource_usage` JSON COMMENT '资源使用情况',
    -- 时间信息
    `created_at` TIMESTAMP NULL COMMENT '创建时间',
    `scheduled_at` TIMESTAMP NULL COMMENT '调度时间',
    `started_at` TIMESTAMP NULL COMMENT '开始时间',
    `finished_at` TIMESTAMP NULL COMMENT '结束时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP NULL COMMENT '删除时间',
    -- 扩展信息
    `extra_info` JSON COMMENT '额外信息',
    
    INDEX `idx_pod_status_job_id` (`job_id`),
    INDEX `idx_pod_status_dag_name` (`dag_name`),
    INDEX `idx_pod_status_cluster` (`cluster`),
    INDEX `idx_pod_status_node_name` (`node_name`),
    INDEX `idx_pod_status_state` (`state`),
    INDEX `idx_pod_status_uuid_retry` (`uuid`, `retry_index`),
    INDEX `idx_pod_status_created` (`created_at`),
    INDEX `idx_pod_status_updated` (`updated_at`),
    INDEX `idx_pod_status_deleted` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Pod状态表';

