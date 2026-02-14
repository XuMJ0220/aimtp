//go:build wireinject
// +build wireinject

package apiserver

import (
	"aimtp/internal/apiserver/biz"
	"aimtp/internal/apiserver/store"
	"aimtp/internal/pkg/server"
	"aimtp/internal/pkg/validation"
	"aimtp/pkg/authz"

	grpcmw "aimtp/internal/pkg/middleware/grpc"
	ginmw "aimtp/internal/pkg/middleware/http"

	"github.com/google/wire"
)

func InitializeWebServer(*Config) (server.Server, error) {
	wire.Build(
		wire.NewSet(NewWebServer, wire.FieldsOf(new(*Config), "ServerMode")),
		wire.Struct(new(ServerConfig), "*"),
		wire.NewSet(store.ProviderSet, biz.ProviderSet),
		ProvideDB, // 提供数据库实例
		validation.ProviderSet,
		wire.NewSet(
			wire.Struct(new(UserRetriever), "*"),
			wire.Bind(new(ginmw.UserRetriever), new(*UserRetriever)),
			wire.Bind(new(grpcmw.UserRetriever), new(*UserRetriever)),
		),
		wire.NewSet(
			authz.ProviderSet,
			wire.Bind(new(grpcmw.Authorizer), new(*authz.Authz)),
		),
	)
	return nil, nil
}
