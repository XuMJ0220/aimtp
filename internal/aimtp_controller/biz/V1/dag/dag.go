package dag

import (
	"aimtp/internal/aimtp_controller/store"
	"aimtp/internal/pkg/k8s"
	v1 "aimtp/pkg/api/aimtp_controller/v1"
	"context"

	argoclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	volcanoclientset "volcano.sh/apis/pkg/client/clientset/versioned"
)

type DAGBiz interface {
	CreateDAG(ctx context.Context, req *v1.CreateDAGRequest) (*v1.CreateDAGResponse, error)
	DAGExpansion
}

type DAGExpansion interface {
}

type dagBiz struct {
	store         store.IStore
	kubeClient    kubernetes.Interface
	volcanoClient volcanoclientset.Interface
	argoClient    argoclientset.Interface
	k8sOpts       *k8s.ConfigOptions
}

// 确保 dagBiz 接口.
var _ DAGBiz = (*dagBiz)(nil)

func New(store store.IStore, kubeClient kubernetes.Interface, volcanoClient volcanoclientset.Interface, argoClient argoclientset.Interface, k8sOpts *k8s.ConfigOptions) *dagBiz {
	return &dagBiz{
		store:         store,
		kubeClient:    kubeClient,
		volcanoClient: volcanoClient,
		argoClient:    argoClient,
		k8sOpts:       k8sOpts,
	}
}
