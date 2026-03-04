package dag

import (
	"aimtp/internal/aimtp_controller/store"

	"k8s.io/client-go/kubernetes"
)

type DAGBiz interface {
	DAGExpansion
}

type DAGExpansion interface {
}

type dagBiz struct {
	store      store.IStore
	kubeClient kubernetes.Interface
}

// 确保 dagBiz 接口.
var _ DAGBiz = (*dagBiz)(nil)

func New(store store.IStore, kubeClient kubernetes.Interface) *dagBiz {
	return &dagBiz{
		store:      store,
		kubeClient: kubeClient,
	}
}
