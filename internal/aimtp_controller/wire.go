//go:build wireinject
// +build wireinject

package aimtp_controller

import (
	"aimtp/internal/aimtp_controller/biz"
	"aimtp/internal/aimtp_controller/pkg/validation"
	"aimtp/internal/aimtp_controller/store"
	"aimtp/internal/pkg/server"

	argoclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"github.com/google/wire"
	"k8s.io/client-go/kubernetes"
	volcanoclientset "volcano.sh/apis/pkg/client/clientset/versioned"
)

func InitializeServer(*Config) (server.Server, error) {
	wire.Build(
		wire.NewSet(NewWebServer, wire.FieldsOf(new(*Config), "ServerMode")),
		wire.Struct(new(ServerConfig), "*"),
		wire.NewSet(store.ProviderSet, biz.ProviderSet),
		ProvideDB,
		validation.ProviderSet,
		ProvideK8sRESTConfig,
		ProvideKubeClient,
		wire.Bind(new(kubernetes.Interface), new(*kubernetes.Clientset)),
		ProvideVolcanoClient,
		wire.Bind(new(volcanoclientset.Interface), new(*volcanoclientset.Clientset)),
		ProvideArgoClient,
		wire.Bind(new(argoclientset.Interface), new(*argoclientset.Clientset)),
	)
	return nil, nil
}
