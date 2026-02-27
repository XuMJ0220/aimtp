package biz

import (
	dagv1 "aimtp/internal/aimtp_server/biz/V1/dag"
	"aimtp/internal/aimtp_server/store"
	"aimtp/internal/pkg/client"

	"github.com/google/wire"
)

// ProviderSet 是一个 Wire 的 Provider 集合，用于声明依赖注入的规则.
// 包含 NewBiz 构造函数，用于生成 biz 实例.
// wire.Bind 用于将接口 IBiz 与具体实现 *biz 绑定，
// 这样依赖 IBiz 的地方会自动注入 *biz 实例.
var ProvicerSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz 定义了业务层需要实现的方法.
type IBiz interface {
	// 获取 DAG 业务接口.
	DAGV1() dagv1.DAGBiz
}

// biz 是 IBiz 的一个具体实现.
type biz struct {
	store             store.IStore
	controllerClients map[string]*client.WorkerClient // 控制器客户端
}

// 确保 biz 实现了 IBiz 接口.
var _ IBiz = (*biz)(nil)

func NewBiz(store store.IStore, controllerClients map[string]*client.WorkerClient) *biz {
	return &biz{
		store:             store,
		controllerClients: controllerClients,
	}
}

// DAGV1 返回一个实现了 DAGBiz 接口的实例.
func (b *biz) DAGV1() dagv1.DAGBiz {
	return dagv1.New(b.store, b.controllerClients)
}
