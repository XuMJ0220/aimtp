package store_test

import (
	"aimtp/internal/aimtp_server/model"
	"aimtp/internal/aimtp_server/store"
	"aimtp/internal/pkg/errno"
	"aimtp/pkg/store/where"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	testDB        *gorm.DB
	testStore     store.IStore
	testContainer testcontainers.Container
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	testDB, testContainer, err = setupMySQLContainer(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup MySQL container: %v", err))
	}

	testStore = store.NewStore(testDB)

	code := m.Run()

	if testContainer != nil {
		testContainer.Terminate(ctx)
	}

	os.Exit(code)
}

func setupMySQLContainer(ctx context.Context) (*gorm.DB, testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.0",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "test1234",
			"MYSQL_DATABASE":      "aimtp",
		},
		WaitingFor: wait.ForLog("port: 3306").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get port: %w", err)
	}

	dsn := fmt.Sprintf("root:test1234@tcp(%s:%s)/aimtp?charset=utf8mb4&parseTime=True&loc=Local", host, port.Port())

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	err = createTableWithSQL(db)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, container, nil
}

func createTableWithSQL(db *gorm.DB) error {
	sql := `
	CREATE TABLE IF NOT EXISTS dag_status_summary (
		dag_id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'DAG自增ID',
		dag_name VARCHAR(255) UNIQUE NOT NULL COMMENT 'DAG名称(唯一)',
		cluster VARCHAR(64) NOT NULL COMMENT '所属集群',
		user_name VARCHAR(128) NOT NULL COMMENT 'DAG所属用户',
		queue_name VARCHAR(128) COMMENT 'DAG所属队列',
		engine VARCHAR(32) DEFAULT 'volcano' COMMENT '执行引擎: volcano/argo',
		state VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'DAG状态',
		creation_status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '创建状态',
		payload MEDIUMTEXT COMMENT 'DAG定义JSON',
		error_msg TEXT COMMENT '创建失败原因',
		retry_count TINYINT DEFAULT 0 COMMENT '重新创建次数',
		max_retries TINYINT DEFAULT 3 COMMENT '最大重新创建次数',
		total_jobs INT DEFAULT 0 COMMENT '总Job数量',
		completed_jobs INT DEFAULT 0 COMMENT '已完成Job数量',
		failed_jobs INT DEFAULT 0 COMMENT '失败Job数量',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
		started_at TIMESTAMP NULL COMMENT 'DAG开始运行时间',
		finished_at TIMESTAMP NULL COMMENT 'DAG结束时间',
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
		deleted_at TIMESTAMP NULL COMMENT '删除时间',
		resource_version VARCHAR(64) COMMENT 'Kubernetes ResourceVersion',
		version BIGINT DEFAULT 0 COMMENT '版本号',
		INDEX idx_cluster (cluster),
		INDEX idx_user (user_name),
		INDEX idx_state (state)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='DAG状态摘要表';
	`
	return db.Exec(sql).Error
}

func cleanupTable(t *testing.T) {
	err := testDB.Exec("TRUNCATE TABLE dag_status_summary").Error
	require.NoError(t, err)
}

func createTestDAG(name string) *model.DagStatusSummaryM {
	return &model.DagStatusSummaryM{
		DagName:        name,
		Cluster:        "test-cluster",
		UserName:       "test-user",
		State:          "pending",
		CreationStatus: "pending",
	}
}

func TestDAGStore_Create(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()
	dag := createTestDAG("test-dag-create")

	err := testStore.DAG().Create(ctx, dag)
	require.NoError(t, err)
	require.NotZero(t, dag.DagID)
}

func TestDAGStore_Create_Duplicate(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()
	dag1 := createTestDAG("test-dag-duplicate")
	dag2 := createTestDAG("test-dag-duplicate")

	err := testStore.DAG().Create(ctx, dag1)
	require.NoError(t, err)

	err = testStore.DAG().Create(ctx, dag2)
	require.Error(t, err)
}

func TestDAGStore_Get(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()
	dag := createTestDAG("test-dag-get")

	err := testStore.DAG().Create(ctx, dag)
	require.NoError(t, err)

	opts := where.NewWhere().F("dag_id", dag.DagID)
	got, err := testStore.DAG().Get(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, dag.DagName, got.DagName)
	require.Equal(t, dag.Cluster, got.Cluster)
	require.Equal(t, dag.UserName, got.UserName)
}

func TestDAGStore_Get_NotFound(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()
	opts := where.NewWhere().F("dag_id", int64(99999))

	_, err := testStore.DAG().Get(ctx, opts)
	require.Error(t, err)
	require.Equal(t, errno.ErrDAGNotFound, err)
}

func TestDAGStore_Update(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()
	dag := createTestDAG("test-dag-update")

	err := testStore.DAG().Create(ctx, dag)
	require.NoError(t, err)

	opts := where.NewWhere().F("dag_id", dag.DagID)
	got, err := testStore.DAG().Get(ctx, opts)
	require.NoError(t, err)

	got.State = "running"
	got.CreationStatus = "created"

	err = testStore.DAG().Update(ctx, got)
	require.NoError(t, err)

	got2, err := testStore.DAG().Get(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, "running", got2.State)
	require.Equal(t, "created", got2.CreationStatus)
}

func TestDAGStore_Delete(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()
	dag := createTestDAG("test-dag-delete")

	err := testStore.DAG().Create(ctx, dag)
	require.NoError(t, err)

	opts := where.NewWhere().F("dag_id", dag.DagID)
	err = testStore.DAG().Delete(ctx, opts)
	require.NoError(t, err)

	_, err = testStore.DAG().Get(ctx, opts)
	require.Error(t, err)
	require.Equal(t, errno.ErrDAGNotFound, err)
}

func TestDAGStore_List(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		dag := createTestDAG(fmt.Sprintf("test-dag-list-%d", i))
		err := testStore.DAG().Create(ctx, dag)
		require.NoError(t, err)
	}

	opts := where.NewWhere().P(1, 10)
	count, list, err := testStore.DAG().List(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, int64(5), count)
	require.Len(t, list, 5)
}

func TestDAGStore_List_WithFilter(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()

	dag1 := createTestDAG("test-dag-filter-1")
	dag1.State = "running"
	err := testStore.DAG().Create(ctx, dag1)
	require.NoError(t, err)

	dag2 := createTestDAG("test-dag-filter-2")
	dag2.State = "pending"
	err = testStore.DAG().Create(ctx, dag2)
	require.NoError(t, err)

	opts := where.NewWhere().F("state", "running")
	count, list, err := testStore.DAG().List(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	require.Len(t, list, 1)
	require.Equal(t, "running", list[0].State)
}

func TestDAGStore_TX_Commit(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()
	dag := createTestDAG("test-dag-tx-commit")

	err := testStore.TX(ctx, func(ctx context.Context) error {
		return testStore.DAG().Create(ctx, dag)
	})
	require.NoError(t, err)

	opts := where.NewWhere().F("dag_id", dag.DagID)
	got, err := testStore.DAG().Get(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, dag.DagName, got.DagName)
}

func TestDAGStore_TX_Rollback(t *testing.T) {
	defer cleanupTable(t)

	ctx := context.Background()
	dag := createTestDAG("test-dag-tx-rollback")

	err := testStore.TX(ctx, func(ctx context.Context) error {
		if err := testStore.DAG().Create(ctx, dag); err != nil {
			return err
		}
		return fmt.Errorf("intentional error for rollback")
	})
	require.Error(t, err)

	opts := where.NewWhere().F("dag_name", "test-dag-tx-rollback")
	_, err = testStore.DAG().Get(ctx, opts)
	require.Error(t, err)
	require.Equal(t, errno.ErrDAGNotFound, err)
}
