package validation

import (
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/known"
	apiv1 "aimtp/pkg/api/aimtp_server/v1"
	genericvalidation "aimtp/pkg/validation"
	"context"
	"regexp"
)

var (
	// K8s资源名称：小写字母、数字、连字符，必须以字母数字开头和结尾
	k8sNamePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	// 镜像名称规则：由字母、数字、分隔符（. _ - / : @）组成
	imagePattern = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9._:/@]*$`)
)

func (v *Validator) ValidateDAGRules() genericvalidation.Rules {

	var (
		validateName func(name string) genericvalidation.ValidatorFunc
	)

	validateName = func(name string) genericvalidation.ValidatorFunc {
		return func(value any) error {
			// 1. 非空检查
			if len(value.(string)) == 0 {
				return errno.ErrInvalidArgument.WithMessage("%s is required", name)
			}

			// 2. 长度检查：1-100
			if len(value.(string)) > 100 {
				return errno.ErrInvalidArgument.WithMessage("%s must not exceed 100 characters", name)
			}

			return nil
		}
	}

	return genericvalidation.Rules{

		"QueueName": validateName("queue_name"),

		"DagName": func(value any) error {
			// 1. 非空检查
			if len(value.(string)) == 0 {
				return errno.ErrInvalidArgument.WithMessage("dag_name is required")
			}

			// 2. 长度检查
			if len(value.(string)) > known.DAGNameMaxLength || len(value.(string)) < known.DAGNameMinLength {
				return errno.ErrInvalidArgument.WithMessage("dag_name must be between 3 and 255 characters")
			}

			// 3. 格式检查
			if !k8sNamePattern.MatchString(value.(string)) {
				return errno.ErrInvalidArgument.WithMessage("dag_name must consist of lower case, alphanumeric, characters or '-', and must start and end with an alphanumeric character")
			}
			return nil
		},

		"UserName": validateName("user_name"),

		"Tasks": func(value any) error {
			tasks, ok := value.([]*apiv1.Task)
			if !ok {
				return errno.ErrInvalidArgument.WithMessage("invalid tasks type")
			}

			if len(tasks) == 0 {
				return errno.ErrInvalidArgument.WithMessage("tasks must be set and cannot be empty")
			}

			if len(tasks) > known.MaxTasksPerDAG {
				return errno.ErrInvalidArgument.WithMessage("tasks must not exceed 100 tasks")
			}

			taskNames := make(map[string]bool)
			// 校验每个任务
			for _, task := range tasks {

				// 1. 校验任务名称
				if task.Name == "" {
					return errno.ErrInvalidArgument.WithMessage("task name must be set and not empty")
				}
				if len(task.Name) > known.TaskNameMaxLength || len(task.Name) < known.TaskNameMinLength {
					return errno.ErrInvalidArgument.WithMessage("task name must be between %d and %d characters", known.TaskNameMinLength, known.TaskNameMaxLength)
				}
				if !k8sNamePattern.MatchString(task.Name) {
					return errno.ErrInvalidArgument.WithMessage("task name must consist of lower case, alphanumeric, characters or '-', and must start and end with an alphanumeric character")
				}

				// 2. 名称的唯一性
				if taskNames[task.Name] {
					return errno.ErrInvalidArgument.WithMessage("task name %s is duplicated", task.Name)
				}
				taskNames[task.Name] = true

				// 3. 校验镜像
				if task.Image == "" {
					return errno.ErrInvalidArgument.WithMessage("task image must be set and not empty")
				}
				if !imagePattern.MatchString(task.Image) {
					return errno.ErrInvalidArgument.WithMessage("task image must consist of alphanumeric characters, '-', '_', '.', '/', ':' or '@'")
				}

				// 4. 校验命令
				if task.Command == nil || task.Command.CommandLine == "" {
					return errno.ErrInvalidArgument.WithMessage("task '%s' command is required", task.Name)
				}
			}

			return nil
		},
	}
}

func (v *Validator) ValidateCreateDAGRequest(ctx context.Context, rq *apiv1.CreateDAGRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateDAGRules())
}
