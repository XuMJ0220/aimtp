package conversion

import (
	v1 "aimtp/pkg/api/aimtp_controller/v1"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ConvertArgs 直接返回 []string，因为现在 proto 已经改成了 repeated string
func ConvertArgs(args []string) []string {
	if args == nil {
		return []string{}
	}
	// 已经是 []string，直接返回（或者做深拷贝）
	return args
}

// ConvertResources 将 Protobuf Resource 转换为 K8s ResourceRequirements
func ConvertResources(res *v1.Resource) corev1.ResourceRequirements {
	if res == nil {
		return corev1.ResourceRequirements{}
	}

	limits := corev1.ResourceList{}
	requests := corev1.ResourceList{}

	// CPU (float32 -> Quantity)
	if res.CPU > 0 {
		cpuQty := resource.MustParse(fmt.Sprintf("%v", res.CPU))
		limits[corev1.ResourceCPU] = cpuQty
		requests[corev1.ResourceCPU] = cpuQty
	}

	// RAM (int64 MB -> Quantity)
	if res.RAM > 0 {
		// 假设 RAM 单位是 MB
		memQty := resource.MustParse(fmt.Sprintf("%dMi", res.RAM))
		limits[corev1.ResourceMemory] = memQty
		requests[corev1.ResourceMemory] = memQty
	}

	// GPU
	if res.GPU > 0 {
		gpuQty := resource.MustParse(fmt.Sprintf("%v", res.GPU))
		limits["nvidia.com/gpu"] = gpuQty
		requests["nvidia.com/gpu"] = gpuQty
	}

	// 其他资源（如 RDMA 等）可在此扩展

	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

// ConvertEnv 将 Protobuf Env map 转换为 K8s EnvVar
func ConvertEnv(env map[string]string) []corev1.EnvVar {
	var result []corev1.EnvVar
	for k, v := range env {
		result = append(result, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return result
}

// ConvertPorts 将 Protobuf ContainerPort 转换为 K8s ContainerPort
func ConvertPorts(ports []*v1.ContainerPort) []corev1.ContainerPort {
	var result []corev1.ContainerPort
	for _, p := range ports {
		protocol := corev1.ProtocolTCP
		if p.Protocol == "UDP" {
			protocol = corev1.ProtocolUDP
		} else if p.Protocol == "SCTP" {
			protocol = corev1.ProtocolSCTP
		}

		result = append(result, corev1.ContainerPort{
			Name:          p.Name,
			ContainerPort: p.ContainerPort,
			Protocol:      protocol,
		})
	}
	return result
}

// ConvertLifecycle 将 Protobuf Lifecycle 转换为 K8s Lifecycle
func ConvertLifecycle(lc *v1.Lifecycle) *corev1.Lifecycle {
	if lc == nil {
		return nil
	}
	// TODO: 实现具体的 Lifecycle 转换逻辑
	// 目前 Protobuf 中的 Handler 定义比较复杂，需要逐一映射到 K8s 的 Exec/HTTPGet
	// 这是一个比较繁琐的工作，需要对照 Handler 结构体
	return nil
}
