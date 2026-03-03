package validation

import (
	"aimtp/internal/aimtp_server/store"

	"github.com/google/wire"
)

// Validator 是验证逻辑的实现结构体.
type Validator struct {
	// 有些复杂的验证逻辑，可能需要直接查询数据库
	// 这里只是一个举例，如果验证时，有其他依赖的客户端/服务/资源等，
	// 都可以一并注入进来
	store store.IStore
}

// ProviderSet 是一个 Wire 的 Provider 集合，用于声明依赖注入的规则.
// 包含 New 构造函数，用于生成 Validator 实例.
var ProviderSet = wire.NewSet(New)

// New 创建一个新的 Validator 实例.
func New(store store.IStore) *Validator {
	return &Validator{store: store}
}
