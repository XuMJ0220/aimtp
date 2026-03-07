package reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aimtp/internal/aimtp_controller/model"
	"aimtp/internal/aimtp_controller/store"
	"aimtp/internal/pkg/log"
	"aimtp/pkg/store/where"

	corev1 "k8s.io/api/core/v1"
)

// PodReconciler 负责同步 Pod 状态到 pod_status 表
type PodReconciler struct {
	cluster string
	store   store.PodStore
}

func NewPodReconciler(cluster string, store store.PodStore) *PodReconciler {
	return &PodReconciler{
		cluster: cluster,
		store:   store,
	}
}

// Reconcile 处理 Pod 的更新
func (r *PodReconciler) Reconcile(ctx context.Context, pod *corev1.Pod) error {
	// 1. 检查是否为 aimtp 管理的 Pod
	// 我们通过 job-name 或 volcano.sh/job-name 等 Label 来判断
	jobID := getJobID(pod)
	if jobID == "" {
		// 不是我们要关心的 Pod
		return nil
	}

	podName := pod.Name
	dagName := pod.Labels["dag-name"]
	// 尝试获取 task-name，Argo 可能叫 workflows.argoproj.io/node-name
	// Volcano 可能不直接提供 task-name，这里作为可选字段

	log.Infow("Reconciling Pod", "pod", podName, "job_id", jobID, "phase", pod.Status.Phase)

	// 2. 查询现有 Pod 记录
	currentPod, err := r.store.Get(ctx, where.NewWhere(
		where.WithQuery("pod_name = ?", podName),
		where.WithQuery("cluster = ?", r.cluster),
	))

	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("get pod failed: %w", err)
	}

	// 3. 构建 Model 对象
	newState := mapPodPhaseToState(pod.Status.Phase)

	// 提取资源使用情况 (Request/Limit)
	resourceUsage := extractResourceUsage(pod)
	resourceJSON, _ := json.Marshal(resourceUsage)
	resourceStr := string(resourceJSON)

	// 提取额外信息
	extraInfo := map[string]interface{}{
		"k8s_phase": string(pod.Status.Phase),
		"qos_class": string(pod.Status.QOSClass),
	}
	extraJSON, _ := json.Marshal(extraInfo)
	extraStr := string(extraJSON)

	// HostIP / PodIP
	hostIP := pod.Status.HostIP
	podIP := pod.Status.PodIP
	nodeName := pod.Spec.NodeName
	reason := string(pod.Status.Reason)
	message := pod.Status.Message

	// 尝试获取退出码
	var exitCode int32
	if len(pod.Status.ContainerStatuses) > 0 {
		state := pod.Status.ContainerStatuses[0].State
		if state.Terminated != nil {
			exitCode = state.Terminated.ExitCode
			if reason == "" {
				reason = state.Terminated.Reason
			}
			if message == "" {
				message = state.Terminated.Message
			}
		}
	}

	if currentPod == nil {
		// 4. Create
		newPod := &model.PodStatusM{
			PodName:       podName,
			JobID:         jobID,
			DagName:       dagName,
			Cluster:       r.cluster,
			Namespace:     &pod.Namespace,
			NodeName:      &nodeName,
			PodIP:         &podIP,
			HostIP:        &hostIP,
			State:         newState,
			Reason:        &reason,
			Message:       &message,
			ExitCode:      &exitCode,
			ResourceUsage: &resourceStr,
			CreatedAt:     &pod.CreationTimestamp.Time,
			ExtraInfo:     &extraStr,
		}

		if pod.Status.StartTime != nil {
			newPod.StartedAt = &pod.Status.StartTime.Time
		}

		if err := r.store.Create(ctx, newPod); err != nil {
			return fmt.Errorf("create pod failed: %w", err)
		}
		log.Infow("Pod created in DB", "pod", podName)

	} else {
		// 5. Update
		// 只有当状态变化或关键信息变更时才更新
		if currentPod.State == newState && *currentPod.PodIP == podIP && *currentPod.HostIP == hostIP {
			return nil
		}

		currentPod.State = newState
		currentPod.PodIP = &podIP
		currentPod.HostIP = &hostIP
		currentPod.NodeName = &nodeName
		currentPod.Reason = &reason
		currentPod.Message = &message
		currentPod.ExitCode = &exitCode
		currentPod.ResourceUsage = &resourceStr
		currentPod.ExtraInfo = &extraStr

		if pod.Status.StartTime != nil && currentPod.StartedAt == nil {
			currentPod.StartedAt = &pod.Status.StartTime.Time
		}

		// 如果是终态，设置 FinishedAt
		if isTerminalState(newState) && currentPod.FinishedAt == nil {
			now := time.Now()
			currentPod.FinishedAt = &now
		}

		if err := r.store.Update(ctx, currentPod); err != nil {
			return fmt.Errorf("update pod failed: %w", err)
		}
		log.Infow("Pod updated in DB", "pod", podName, "new_state", newState)
	}

	return nil
}

func getJobID(pod *corev1.Pod) string {
	// Volcano Job
	if jobName, ok := pod.Labels["volcano.sh/job-name"]; ok {
		return jobName
	}
	// Argo Workflow
	if wfName, ok := pod.Labels["workflows.argoproj.io/workflow"]; ok {
		return wfName
	}
	// 普通 Job
	if jobName, ok := pod.Labels["job-name"]; ok {
		return jobName
	}
	// 自定义 Label
	if jobID, ok := pod.Labels["job-id"]; ok {
		return jobID
	}
	return ""
}

func mapPodPhaseToState(phase corev1.PodPhase) string {
	switch phase {
	case corev1.PodPending:
		return "Pending"
	case corev1.PodRunning:
		return "Running"
	case corev1.PodSucceeded:
		return "Succeeded"
	case corev1.PodFailed:
		return "Failed"
	default:
		return string(phase)
	}
}

func extractResourceUsage(pod *corev1.Pod) map[string]string {
	usage := make(map[string]string)
	for _, container := range pod.Spec.Containers {
		for k, v := range container.Resources.Requests {
			usage["request_"+k.String()] = v.String()
		}
		for k, v := range container.Resources.Limits {
			usage["limit_"+k.String()] = v.String()
		}
	}
	return usage
}

// Delete 处理 Pod 的删除
func (r *PodReconciler) Delete(ctx context.Context, namespace, name string) error {
	podName := name
	log.Infow("Reconciling Pod Deletion", "pod", podName, "cluster", r.cluster)

	// 直接调用 Delete，GORM 会根据 Model 中的 DeletedAt 字段自动处理软删除
	if err := r.store.Delete(ctx, where.NewWhere(
		where.WithQuery("pod_name = ?", podName),
		where.WithQuery("cluster = ?", r.cluster),
	)); err != nil {
		return fmt.Errorf("delete pod failed: %w", err)
	}

	log.Infow("Pod deleted from DB (soft delete)", "pod", podName)
	return nil
}
