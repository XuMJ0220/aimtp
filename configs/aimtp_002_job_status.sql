-- Active: 1772191086950@@192.168.124.9@3306@aimtp
-- =============================================
-- Local: job_status 表
-- 描述: Job 状态表（DAG → Job → Pod 三层架构）
-- =============================================

CREATE TABLE IF NOT EXISTS `job_status` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `job_id` VARCHAR(255) UNIQUE NOT NULL COMMENT 'Job全局唯一ID',
    -- 所属 DAG
    `dag_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'DAG ID（关联 dag_status.dag_id）',
    `dag_name` VARCHAR(255) NOT NULL COMMENT 'DAG名称',
    `job_name` VARCHAR(255) NOT NULL COMMENT 'Job名称（任务名）',
    `cluster` VARCHAR(64) NOT NULL COMMENT '集群名称',
    -- Job 类型与引擎
    `job_type` VARCHAR(32) NOT NULL DEFAULT 'volcano_job' COMMENT 'Job类型：volcano_job',
    `engine` VARCHAR(32) NOT NULL DEFAULT 'volcano' COMMENT '执行引擎：volcano',
    -- 关联 Kubernetes 资源
    `vj_name` VARCHAR(255) COMMENT 'Volcano Job 名称',
    `workflow_name` VARCHAR(255) COMMENT 'Argo Workflow 名称（多Job DAG）',
    `node_id` VARCHAR(255) COMMENT 'Argo Node ID',
    `namespace` VARCHAR(128) DEFAULT 'default' COMMENT '命名空间',
    -- 状态信息
    `state` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'Job状态：pending/queuing/running/succeeded/failed/terminating',
    `message` TEXT COMMENT '详细消息',
    `reason` VARCHAR(512) COMMENT '原因',
    -- Pod 统计
    `expected_pod_count` INT DEFAULT 1 COMMENT '期望Pod数量',
    `running_pod_count` INT DEFAULT 0 COMMENT '运行中Pod数量',
    `succeeded_pod_count` INT DEFAULT 0 COMMENT '成功Pod数量',
    `failed_pod_count` INT DEFAULT 0 COMMENT '失败Pod数量',
    -- 执行信息
    `retry_count` INT DEFAULT 0 COMMENT '重试次数',
    `priority` INT DEFAULT 0 COMMENT '优先级',
    -- 时间信息
    `created_at` TIMESTAMP NULL COMMENT '创建时间',
    `started_at` TIMESTAMP NULL COMMENT '开始时间',
    `finished_at` TIMESTAMP NULL COMMENT '结束时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` TIMESTAMP COMMENT '删除时间',
    `job_home` VARCHAR(255) COMMENT 'Job归档路径',
    `duration` INT COMMENT '运行时长（秒）',
    -- 扩展信息
    `extra_info` JSON COMMENT '额外信息',
    
    INDEX `idx_job_status_dag_id` (`dag_id`),
    INDEX `idx_job_status_dag_name` (`dag_name`),
    INDEX `idx_job_status_job_name` (`job_name`),
    INDEX `idx_job_status_cluster` (`cluster`),
    INDEX `idx_job_status_state` (`state`),
    INDEX `idx_job_status_vj_name` (`vj_name`),
    INDEX `idx_job_status_workflow_node` (`workflow_name`, `node_id`),
    INDEX `idx_job_status_created` (`created_at`),
    INDEX `idx_job_status_updated` (`updated_at`),
    INDEX `idx_job_status_deleted` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Job状态表';

