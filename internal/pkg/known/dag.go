package known

const (
	// MaxPayloadSize 定义了 DAG  payload 的最大大小.
	// 用于限制 DAG payload 的大小，防止恶意请求占用过多资源.
	MaxPayloadSize = 1*1024*1024 // 1MB

	// DAGInitState 定义了 DAG 初始化状态.
	DAGInitState = "pending"
	// DAGInitCreationStatus 定义了 DAG 初始化创建状态.
	DAGInitCreationStatus = "pending"
	// DAGMaxRetries 定义了 DAG 最大重试次数.
	DAGMaxRetries int32 = 3

	// DAGNameMinLength	定义了 DAG 名称的最小长度.
	DAGNameMinLength = 3
	// DAGNameMaxLength 定义了 DAG 名称的最大长度.
	DAGNameMaxLength = 255
	// TaskNameMinLength
	TaskNameMinLength = 1
	// TaskNameMaxLength
	TaskNameMaxLength = 63

	// MaxTasksPerDAG 定义了每个 DAG 允许的最大任务数.
	MaxTasksPerDAG = 100
)
