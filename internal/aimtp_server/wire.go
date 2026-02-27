//go:build wireinject
// +build wireinject

package aimtp_server

import (
	"aimtp/internal/aimtp_server/biz"
	"aimtp/internal/aimtp_server/pkg/validation"
	"aimtp/internal/aimtp_server/store"
	"aimtp/internal/pkg/client"
	"aimtp/internal/pkg/server"

	"github.com/google/wire"
)

func InitializeServer(*Config) (server.Server, error) {
	wire.Build(
		wire.NewSet(NewWebServer, wire.FieldsOf(new(*Config), "ServerMode")),
		wire.Struct(new(ServerConfig), "*"),
		wire.NewSet(store.ProviderSet, biz.ProvicerSet),
		ProvideDB,
		validation.ProviderSet,
		wire.FieldsOf(new(*Config), "ControllerClusters"),
		client.NewControllerClients,
	)
	return nil, nil
}
