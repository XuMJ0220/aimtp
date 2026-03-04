package k8s

import (
	argoclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	volcanoclientset "volcano.sh/apis/pkg/client/clientset/versioned"
)

func NewKubeClient(cfg *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(cfg)
}

func NewVolcanoClient(cfg *rest.Config) (*volcanoclientset.Clientset, error) {
	return volcanoclientset.NewForConfig(cfg)
}

func NewArgoClient(cfg *rest.Config) (*argoclientset.Clientset, error) {
	return argoclientset.NewForConfig(cfg)
}

