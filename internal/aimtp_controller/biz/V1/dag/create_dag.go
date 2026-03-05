package dag

import (
	"aimtp/internal/aimtp_controller/model"
	"aimtp/internal/aimtp_controller/pkg/conversion"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/log"
	v1 "aimtp/pkg/api/aimtp_controller/v1"
	"context"
	"encoding/json"
	"fmt"
	"time"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/kballard/go-shellquote"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vcbatch "volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

const (
	EngineVolcano = "volcano"
	EngineArgo    = "argo"

	JobTypeVolcano = "volcano_job"
	JobTypeArgo    = "argo_workflow"
)

func (b *dagBiz) CreateDAG(ctx context.Context, req *v1.CreateDAGRequest) (*v1.CreateDAGResponse, error) {
	// 根据任务数量和依赖决定创建类型
	// 策略：单任务(无依赖) -> Volcano Job; 多任务 -> Argo Workflow
	var err error
	// 判断是否有依赖关系
	hasDependencies := false
	if len(req.Dependencies) > 0 {
		hasDependencies = true
	}

	if len(req.Tasks) == 1 && !hasDependencies {
		err = b.createSingleTask(ctx, req)
	} else {
		err = b.createWorkflow(ctx, req)
	}

	if err != nil {
		log.Errorw("Failed to create DAG", "err", err, "dag_name", req.DagName)
		return nil, err
	}

	return &v1.CreateDAGResponse{}, nil
}

// createSingleTask 创建单任务（直接创建 Volcano Job）
func (b *dagBiz) createSingleTask(ctx context.Context, req *v1.CreateDAGRequest) error {
	namespace := req.QueueName
	if namespace == "" {
		namespace = "default"
	}

	task := req.Tasks[0]
	// Job名称格式：aimtp-{task_name}
	jobName := fmt.Sprintf("aimtp-%s", task.Name)
	// 任务类型
	jobType := task.Type
	if jobType == "" {
		jobType = JobTypeVolcano
	}
	// Engine 类型
	engine := EngineVolcano
	if jobType == JobTypeArgo {
		engine = EngineArgo
	}

	// 1. 预写 DB (JobStatus)
	now := time.Now()
	jobStatus := &model.JobStatusM{
		JobID:     jobName,
		DagName:   req.DagName,
		Cluster:   req.Cluster,
		JobType:   jobType,
		Engine:    engine,
		State:     "pending",
		CreatedAt: &now,
		UpdatedAt: now,
		VjName:    &jobName,
		Namespace: &namespace,
		JobName:   task.Name,
	}

	if err := b.store.Job().Create(ctx, jobStatus); err != nil {
		return err
	}

	// 2. 构造 Volcano Job CRD
	vcJob := &vcbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"aimtp.io/dag-name": req.DagName,
				"aimtp.io/user":     req.UserName,
				"aimtp.io/job-name": task.Name,
			},
		},
		Spec: vcbatch.JobSpec{
			SchedulerName: "volcano",
			Queue:         req.QueueName,
			Tasks: []vcbatch.TaskSpec{
				{
					Name:     req.DagName, // TaskSpec.Name = req.DagName，用于生成 Pod Name
					Replicas: int32(task.InstanceCount),
					Template: b.buildPodSpec(task, req),
				},
			},
		},
	}

	// 3. 调用 Client 创建
	_, err := b.volcanoClient.BatchV1alpha1().Jobs(vcJob.Namespace).Create(ctx, vcJob, metav1.CreateOptions{})
	if err != nil {
		jobStatus.State = "failed"
		msg := err.Error()
		jobStatus.Message = &msg
		_ = b.store.Job().Update(ctx, jobStatus)
		return errno.ErrCreateDAGFailed.WithMessage("failed to create volcano job: %s", err.Error())
	}

	return nil
}

// createWorkflow 创建多任务工作流（Argo Workflow）
func (b *dagBiz) createWorkflow(ctx context.Context, req *v1.CreateDAGRequest) error {
	namespace := req.QueueName
	if namespace == "" {
		namespace = "default"
	}

	// 1. 预写 DB (为每个 Task 创建 JobStatus)
	now := time.Now()
	for _, task := range req.Tasks {
		vjJobName := fmt.Sprintf("aimtp-job-%s-%s", req.DagName, task.Name)
		jobStatus := &model.JobStatusM{
			JobID:        vjJobName,
			DagName:      req.DagName,
			Cluster:      req.Cluster,
			JobType:      JobTypeArgo, // 修改为 Argo
			Engine:       EngineArgo,  // 修改为 Argo
			State:        "pending",
			CreatedAt:    &now,
			UpdatedAt:    now,
			VjName:       &vjJobName,
			Namespace:    &namespace,
			WorkflowName: &req.DagName,
			JobName:      task.Name, // 补充 JobName
		}
		if err := b.store.Job().Create(ctx, jobStatus); err != nil {
			return err // TODO: 考虑回滚已创建的记录
		}
	}

	// 2. 构造 Argo Workflow CRD
	wf := &wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.DagName,
			Namespace: namespace,
			Labels: map[string]string{
				"aimtp.io/dag-name": req.DagName,
			},
		},
		Spec: wfv1.WorkflowSpec{
			Entrypoint: "main",
			Templates: []wfv1.Template{
				{
					Name: "main",
					DAG:  &wfv1.DAGTemplate{},
				},
			},
		},
	}

	// 构建 DAG Template
	dagTask := &wfv1.DAGTemplate{
		Tasks: make([]wfv1.DAGTask, 0, len(req.Tasks)),
	}

	for _, task := range req.Tasks {
		// 使用 Resource Template 创建 Volcano Job
		tmplName := task.Name + "-tmpl"
		jobName := fmt.Sprintf("aimtp-job-%s-%s", req.DagName, task.Name)

		// 构造内嵌的 Volcano Job YAML
		// 注意：这里需要构造完整的 Volcano Job 结构
		// 为了简化，这里先只填充最基本的结构
		vcJob := &vcbatch.Job{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "batch.volcano.sh/v1alpha1",
				Kind:       "Job",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: namespace,
				Labels: map[string]string{
					"aimtp.io/dag-name":  req.DagName,
					"aimtp.io/task-name": task.Name,
				},
			},
			Spec: vcbatch.JobSpec{
				SchedulerName: "volcano",
				Queue:         req.QueueName,
				Tasks: []vcbatch.TaskSpec{
					{
						Name:     task.Name,
						Replicas: int32(task.InstanceCount),
						Template: b.buildPodSpec(task, req),
					},
				},
			},
		}

		// 将 vcJob 转为 manifest string
		// 实际开发中可能需要 yaml.Marshal 或直接构造 ResourceTemplate
		// Argo 支持直接嵌入 K8s 资源
		wf.Spec.Templates = append(wf.Spec.Templates, wfv1.Template{
			Name: tmplName,
			Resource: &wfv1.ResourceTemplate{
				Action:   "create",
				Manifest: mustMarshal(vcJob), // 辅助函数
			},
		})

		// 添加 DAG Task
		deps := []string{}
		if req.Dependencies != nil {
			if list, ok := req.Dependencies[task.Name]; ok {
				deps = list.Items
			}
		}

		wfTask := wfv1.DAGTask{
			Name:         task.Name,
			Template:     tmplName,
			Dependencies: deps,
		}
		dagTask.Tasks = append(dagTask.Tasks, wfTask)
	}
	wf.Spec.Templates[0].DAG = dagTask

	// 3. 调用 Client 创建
	_, err := b.argoClient.ArgoprojV1alpha1().Workflows(wf.Namespace).Create(ctx, wf, metav1.CreateOptions{})
	if err != nil {
		// 创建失败，更新所有相关 Job 的状态为 Failed
		errMsg := err.Error()
		for _, task := range req.Tasks {
			jobName := fmt.Sprintf("aimtp-job-%s-%s", req.DagName, task.Name)
			// 这里需要查出来再更新，或者直接构造一个只包含 ID 和状态的结构体去更新
			// 假设 store.Job().Update 支持部分更新（根据 ID）
			jobStatus := &model.JobStatusM{
				JobID:   jobName,
				State:   "failed",
				Message: &errMsg,
			}
			// 忽略更新错误，尽力而为
			_ = b.store.Job().Update(ctx, jobStatus)
		}
		return errno.ErrCreateDAGFailed.WithMessage("failed to create argo workflow: %s", err.Error())
	}

	return nil
}

func (b *dagBiz) buildPodSpec(task *v1.Task, req *v1.CreateDAGRequest) corev1.PodTemplateSpec {
	// 构建挂载卷
	vols, vms, err := b.buildMounts(task, req)
	if err != nil {
		log.Errorw("Failed to build mounts", "err", err, "task", task.Name)
		// 如果挂载失败，暂时只记录日志，不阻断（或者根据业务需求 panic/error）
	}

	// 解析命令行
	var command []string
	var args []string
	if task.GetCommand().GetCommandLine() != "" {
		// 使用 shellquote 解析命令行字符串
		parts, err := shellquote.Split(task.GetCommand().GetCommandLine())
		if err != nil {
			log.Errorw("Failed to split command line", "err", err, "command", task.GetCommand().GetCommandLine())
			// 降级处理：如果不符合 shell 规则，则整体作为一个参数
			command = []string{task.GetCommand().GetCommandLine()}
		} else {
			if len(parts) > 0 {
				command = []string{parts[0]}
				if len(parts) > 1 {
					args = parts[1:]
				}
			}
		}
	}

	// 如果 Task 中显式定义了 Args，则追加到解析出的 args 后面（或者覆盖，取决于业务约定）
	// 这里假设 Task.Args 是额外的参数
	extraArgs := conversion.ConvertArgs(task.GetCommand().GetArgs())
	args = append(args, extraArgs...)

	return corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:         task.Name,
					Image:        task.Image,
					Command:      command,
					Args:         args,
					Resources:    conversion.ConvertResources(task.Resources),
					Env:          conversion.ConvertEnv(task.Env),
					Ports:        conversion.ConvertPorts(task.Ports),
					Lifecycle:    conversion.ConvertLifecycle(task.Lifecycle),
					VolumeMounts: vms,
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
			Volumes:       vols,
		},
	}
}

// buildMounts 构建 Volumes 和 VolumeMounts
func (b *dagBiz) buildMounts(task *v1.Task, req *v1.CreateDAGRequest) ([]corev1.Volume, []corev1.VolumeMount, error) {
	var vols []corev1.Volume
	var vms []corev1.VolumeMount

	// 获取存储配置
	storageType := b.k8sOpts.Storage.Type
	hostPathPrefix := b.k8sOpts.Storage.HostPathPrefix
	// pvcName := b.k8sOpts.Storage.PVCName // 暂时未用到

	if storageType == "" {
		storageType = "hostPath" // 默认
	}
	if hostPathPrefix == "" {
		hostPathPrefix = "/data/" // 默认
	}

	// 1. 基础系统挂载 (参照 HCT)
	// /share -> {root}/share
	vols, vms = appendMount(vols, vms, "share-dir", hostPathPrefix+"share", "/share", storageType)

	// /running_root -> {root}/plat_gpu
	vols, vms = appendMount(vols, vms, "computing-dfs-root", hostPathPrefix+"plat_gpu", "/running_root", storageType)

	// /cluster_tmp -> {root}/cluster_tmp
	vols, vms = appendMount(vols, vms, "dfs-cluster-tmp", hostPathPrefix+"cluster_tmp", "/cluster_tmp", storageType)

	// /dev/shm (Memory)
	vols = append(vols, corev1.Volume{
		Name: "shmem",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumMemory,
			},
		},
	})
	vms = append(vms, corev1.VolumeMount{Name: "shmem", MountPath: "/dev/shm"})

	// 2. 用户相关挂载
	// 挂载用户 Home 目录: {hostPath}/users/{username} -> /users/{username} 和 /cluster_home
	if req.UserName != "" {
		userHomeHostPath := fmt.Sprintf("%susers/%s", hostPathPrefix, req.UserName)
		vols, vms = appendMount(vols, vms, "dfs-user-home", userHomeHostPath, "/cluster_home", storageType)

		// 同时也挂载到容器内的标准 /users/{username} 路径
		vms = append(vms, corev1.VolumeMount{
			Name:      "dfs-user-home",
			MountPath: fmt.Sprintf("/users/%s", req.UserName),
		})
	}

	// 3. Task 特定的挂载 (如有)
	// TODO: 处理 req.MountConfig (如果 CreateDAGRequest 中有定义 MountConfig 字段的话)
	// 目前 Protobuf 定义中 CreateDAGRequest 没有 MountConfig 字段，只有 Task.Env 等
	// 如果需要支持自定义挂载，需要在 Protobuf 中增加 MountConfig 定义

	return vols, vms, nil
}

func appendMount(vols []corev1.Volume, vms []corev1.VolumeMount, name, hostPath, mountPath, storageType string) ([]corev1.Volume, []corev1.VolumeMount) {
	if storageType == "hostPath" {
		vols = append(vols, corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostPath,
				},
			},
		})
	} else if storageType == "pvc" {
		// PVC 模式下，通常是挂载 PVC 的子路径
		// 这里简化处理，假设所有数据都在一个 PVC 下
		// 需要外部传入 PVC Name
		// vols = append(vols, corev1.Volume{Name: "data-pvc", PersistentVolumeClaim: ...})
		// vms = append(vms, SubPath: strings.TrimPrefix(hostPath, prefix))
	} else {
		// emptyDir
		vols = append(vols, corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	vms = append(vms, corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	})
	return vols, vms
}

func mustMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal object: %v", err))
	}
	return string(b)
}
