package dag

import (
	"aimtp/internal/aimtp_controller/store"

	argoclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	volcanoclientset "volcano.sh/apis/pkg/client/clientset/versioned"
)

type DAGBiz interface {
	DAGExpansion
}

type DAGExpansion interface {
}

type dagBiz struct {
	store         store.IStore
	kubeClient    kubernetes.Interface
	volcanoClient volcanoclientset.Interface
	argoClient    argoclientset.Interface
}

// 确保 dagBiz 接口.
var _ DAGBiz = (*dagBiz)(nil)

func New(store store.IStore, kubeClient kubernetes.Interface, volcanoClient volcanoclientset.Interface, argoClient argoclientset.Interface) *dagBiz {
	return &dagBiz{
		store:         store,
		kubeClient:    kubeClient,
		volcanoClient: volcanoClient,
		argoClient:    argoClient,
	}
}
