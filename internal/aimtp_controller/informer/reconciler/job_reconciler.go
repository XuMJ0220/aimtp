package reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aimtp/internal/aimtp_controller/store"
	"aimtp/internal/pkg/log"
	"aimtp/pkg/store/where"

	volcanov1 "volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

// JobReconciler 负责同步 Volcano Job 状态到 job_status 表
type JobReconciler struct {
	cluster  string
	store    store.JobStore
	podStore store.PodStore
}

func NewJobReconciler(cluster string, store store.JobStore, podStore store.PodStore) *JobReconciler {
	return &JobReconciler{
		cluster:  cluster,
		store:    store,
		podStore: podStore,
	}
}

// Reconcile 处理 Volcano Job 的更新
func (r *JobReconciler) Reconcile(ctx context.Context, vj *volcanov1.Job) error {
	jobID := vj.Name
	dagName := vj.Labels["dag-name"]
	taskName := vj.Labels["task-name"]

	log.Infow("Reconciling Job", "job_id", jobID, "cluster", r.cluster, "phase", vj.Status.State.Phase)

	// 1. 查询现有 Job 记录
	// 必须加上 Cluster 条件，防止多集群 Job ID 冲突
	currentJob, err := r.store.Get(ctx, where.NewWhere(
		where.WithQuery("job_id = ?", jobID),
		where.WithQuery("cluster = ?", r.cluster),
	))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Job 不存在，可能是被手动删除了，或者创建流程还没写入
			// 这里我们选择忽略，因为我们只负责更新已存在的 Job 状态
			log.Warnw("Job not found in DB, skipping reconcile", "job_id", jobID)
			return nil
		}
		return fmt.Errorf("get job failed: %w", err)
	}

	// 2. 映射状态
	newState := mapVJPhaseToState(vj.Status.State.Phase)
	if newState == "" {
		log.Warnw("Unknown Volcano Phase, skipping", "phase", vj.Status.State.Phase)
		return nil
	}

	// 3. 检查状态是否需要更新
	// 如果状态没有变化，且没有新的时间戳，则跳过
	if currentJob.State == newState {
		return nil
	}

	// 4. 更新字段
	currentJob.State = newState
	currentJob.Message = &vj.Status.State.Message
	reason := string(vj.Status.State.Reason)
	currentJob.Reason = &reason

	// 更新时间戳
	if !vj.Status.State.LastTransitionTime.IsZero() && currentJob.StartedAt == nil {
		t := vj.Status.State.LastTransitionTime.Time
		currentJob.StartedAt = &t
	}

	if isTerminalState(newState) && currentJob.FinishedAt == nil {
		now := time.Now()
		currentJob.FinishedAt = &now

		// 计算运行时长
		if currentJob.StartedAt != nil {
			duration := int32(now.Sub(*currentJob.StartedAt).Seconds())
			currentJob.Duration = &duration
		}
	}

	// 记录原始信息
	extraInfo := map[string]interface{}{
		"k8s_phase":  string(vj.Status.State.Phase),
		"k8s_reason": string(vj.Status.State.Reason),
		"dag_name":   dagName,
		"task_name":  taskName,
	}
	if currentJob.ExtraInfo != nil {
		var oldExtra map[string]interface{}
		if err := json.Unmarshal([]byte(*currentJob.ExtraInfo), &oldExtra); err == nil {
			for k, v := range oldExtra {
				extraInfo[k] = v
			}
		}
	}
	extraJSON, _ := json.Marshal(extraInfo)
	extraStr := string(extraJSON)
	currentJob.ExtraInfo = &extraStr

	// 5. 保存更新
	if err := r.store.Update(ctx, currentJob); err != nil {
		return fmt.Errorf("update job failed: %w", err)
	}

	log.Infow("Job reconciled successfully", "job_id", jobID, "new_state", newState)
	return nil
}

// Delete 处理 Volcano Job 的删除
func (r *JobReconciler) Delete(ctx context.Context, namespace, name string) error {
	jobID := name
	log.Infow("Reconciling Job Deletion", "job_id", jobID, "cluster", r.cluster)

	// 1. 标记为已删除
	// 如果 Model 使用了 gorm.DeletedAt，可以直接调用 Delete
	// 如果没有，需要手动设置 DeletedAt 字段
	// 假设 store.Delete 实现了软删除（基于 GORM）
	if err := r.store.Delete(ctx, where.NewWhere(
		where.WithQuery("job_id = ?", jobID),
		where.WithQuery("cluster = ?", r.cluster),
	)); err != nil {
		return fmt.Errorf("delete job failed: %w", err)
	}

	log.Infow("Job deleted from DB (soft delete)", "job_id", jobID)
	return nil
}

// mapVJPhaseToState 将 Volcano Phase 映射为系统内部状态
func mapVJPhaseToState(phase volcanov1.JobPhase) string {
	switch phase {
	case volcanov1.Pending:
		return "Pending"
	case volcanov1.Running, volcanov1.Restarting:
		return "Running"
	case volcanov1.Completed:
		return "Succeeded" // 注意：这里我们统一叫 Succeeded，对应 HCT 的 Completed
	case volcanov1.Failed, volcanov1.Terminated, volcanov1.Aborted:
		return "Failed"
	default:
		return string(phase)
	}
}

func isTerminalState(state string) bool {
	return state == "Succeeded" || state == "Failed"
}
