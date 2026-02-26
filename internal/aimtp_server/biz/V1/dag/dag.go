package dag

import (
	"aimtp/internal/aimtp_server/model"
	"aimtp/internal/aimtp_server/store"
	"aimtp/internal/pkg/client"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/known"
	"aimtp/internal/pkg/log"
	apiv1 "aimtp/pkg/api/aimtp_server/v1"
	"aimtp/pkg/store/where"
	"context"
)

type DAGBiz interface {
	CreateDAG(ctx context.Context, rq *apiv1.CreateDAGRequest) (*apiv1.CreateDAGResponse, error)

	DAGExpansion
}

type DAGExpansion interface {
}

type dagBiz struct {
	store             store.IStore
	controllerClients map[string]*client.WorkerClient // 控制器客户端
}

// 确保 dagBiz 接口.
var _ DAGBiz = (*dagBiz)(nil)

func New(store store.IStore, controllerClients map[string]*client.WorkerClient) *dagBiz {
	return &dagBiz{
		store:             store,
		controllerClients: controllerClients,
	}
}

func (b *dagBiz) CreateDAG(ctx context.Context, rq *apiv1.CreateDAGRequest) (*apiv1.CreateDAGResponse, error) {

	dagName := rq.GetDagName()
	cluster := rq.GetCluster()
	totalJobs := int32(len(rq.Tasks))
	maxRetries := known.DAGMaxRetries
	retryCount := int32(0)

	// 检查 DAG 是否存在
	_, err := b.store.DAG().Get(ctx, where.F("dag_name", dagName))
	if err == nil {
		return nil, errno.ErrDAGAlreadyExist.WithMessage("DAG %s already exist.", dagName)
	}

	// 选择目标集群
	cluster, err = b.selectCluster(ctx, cluster)
	if err != nil {
		return nil, err
	}

	// 序列化并验证 DAG payload
	payload, err := b.serializeAndValidatePayload(rq)
	if err != nil {
		return nil, err
	}

	// 写入数据库
	dagStatusSummaryM := &model.DagStatusSummaryM{
		DagName:        dagName,
		Cluster:        cluster,
		UserName:       rq.GetUserName(),
		QueueName:      &rq.QueueName,
		Engine:         &rq.Engine,
		State:          known.DAGInitState,
		CreationStatus: known.DAGInitCreationStatus,
		Payload:        &payload,
		RetryCount:     &retryCount,
		MaxRetries:     &maxRetries,
		TotalJobs:      &totalJobs,
	}

	if err := b.store.DAG().Create(ctx, dagStatusSummaryM); err != nil {
		log.Errorw("Failed to create DAG", "err", err)
		return nil, errno.ErrCreateDAGFailed.WithMessage("Failed to create DAG, err: %s", err.Error())
	}

	log.Infow("DAG created successfully", "DAGName", dagName, "cluster", cluster, "tasks", totalJobs)
	return &apiv1.CreateDAGResponse{}, nil
}
