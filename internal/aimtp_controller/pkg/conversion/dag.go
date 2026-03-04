package conversion

import (
	v1 "aimtp/pkg/api/aimtp_controller/v1"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ConvertArgs 将 structpb.Struct 转换为字符串数组
// 假设 args 是一个 JSON 对象，我们将其 key=value 或其他形式转换为 string list
// 这里的具体策略取决于业务约定。
// 常见策略：
// 1. 如果是 {"-f": "file", "--verbose": true} -> ["-f", "file", "--verbose"]
// 2. 如果是 ordered list，structpb 可能不合适，应该用 ListValue
//
// 鉴于 HCT 的历史代码，这里可能需要根据实际情况调整。
// 目前简单实现为：将整个 struct 转为 JSON 字符串作为第一个参数，或者忽略。
//
// UPDATE: 根据常见 K8s 实践，Args 通常是列表。但 Protobuf 定义为 Struct。
// 我们暂时假设它是一个 map，转换为 key=value 形式，或者根据具体业务逻辑处理。
// 为了跑通流程，这里先做一个简单的扁平化处理：key=value
func ConvertArgs(args *structpb.Struct) []string {
	if args == nil {
		return []string{}
	}

	var result []string
	for k, v := range args.Fields {
		// 简单处理 string, number, bool
		valStr := ""
		switch x := v.Kind.(type) {
		case *structpb.Value_StringValue:
			valStr = x.StringValue
		case *structpb.Value_NumberValue:
			valStr = fmt.Sprintf("%v", x.NumberValue)
		case *structpb.Value_BoolValue:
			valStr = fmt.Sprintf("%v", x.BoolValue)
		default:
			// 复杂类型转 JSON
			b, _ := json.Marshal(v)
			valStr = string(b)
		}

		// 假设 args 是命名参数，如 --key value
		// 这里假设 key 就是参数名
		if k != "" {
			result = append(result, k)
		}
		if valStr != "" {
			result = append(result, valStr)
		}
	}
	return result
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
