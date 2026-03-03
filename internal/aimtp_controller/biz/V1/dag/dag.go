package dag

import (
	"aimtp/internal/aimtp_controller/store"
)

type DAGBiz interface {
	DAGExpansion
}

type DAGExpansion interface {
}

type dagBiz struct {
	store store.IStore
}

// 确保 dagBiz 接口.
var _ DAGBiz = (*dagBiz)(nil)

func New(store store.IStore) *dagBiz {
	return &dagBiz{
		store: store,
	}
}
