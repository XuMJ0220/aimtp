package dag

import (
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/known"
	"aimtp/internal/pkg/log"
	apiv1 "aimtp/pkg/api/aimtp_server/v1"
	"context"
	"encoding/json"
	"sort"
	"strings"
)

// selectCluster 选择目标集群
func (b *dagBiz) selectCluster(ctx context.Context, cluster string) (string, error) {
	// 如果指定了集群，使用指定的集群
	if c := strings.TrimSpace(cluster); c != "" {
		// 检查指定的集群是否存在
		if _, ok := b.controllerClients[c]; !ok {
			return "", errno.ErrClusterNotFound.WithMessage("Cluster %s not found", c)
		}

		// 检查指定的集群是否健康
		if err := b.controllerClients[c].HealthCheck(ctx); err != nil {
			return "", errno.ErrClusterUnhealthy.WithMessage("Cluster %s is unhealthy, err: %s", c, err.Error())
		}

		return c, nil
	}

	// 如果没有集群可以选择
	if len(b.controllerClients) == 0 {
		return "", errno.ErrClusterNotFound.WithMessage("No cluster available")
	}

	// 获取集群的名字
	clusters := make([]string, 0, len(b.controllerClients))
	for cluster := range b.controllerClients {
		clusters = append(clusters, cluster)
	}
	sort.Strings(clusters)
	// 选择一个健康的集群
	for _, cluster := range clusters {
		client := b.controllerClients[cluster]
		if err := client.HealthCheck(ctx); err != nil {
			log.Warnw("Cluster is unhealthy", "cluster", cluster, "err", err)
			continue
		}

		log.Infow("Selected cluster", "cluster", cluster)
		return cluster, nil
	}

	// 如果没有健康的集群
	log.Errorw("All clusters are unhealthy", "clusters", clusters)
	return "", errno.ErrClusterUnhealthy.WithMessage("All clusters are unhealthy")
}

// serializeAndValidatePayload 序列化并验证 DAG payload
func (b *dagBiz) serializeAndValidatePayload(rq *apiv1.CreateDAGRequest) (string, error) {
	payloadBytes, err := json.Marshal(rq)
	if err != nil {
		return "", errno.ErrSerializeDAGPayload.WithMessage("Failed to serialize DAG payload, err: %s", err.Error())
	}

	if len(payloadBytes) > known.MaxPayloadSize {
		return "", errno.ErrSerializeDAGPayload.WithMessage("DAG payload size exceeds the maximum limit of %d bytes", known.MaxPayloadSize)
	}

	return string(payloadBytes), nil
}
