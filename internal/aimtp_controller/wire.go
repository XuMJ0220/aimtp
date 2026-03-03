//go:build wireinject
// +build wireinject

package aimtp_controller

import (
	"aimtp/internal/aimtp_controller/biz"
	"aimtp/internal/aimtp_controller/pkg/validation"
	"aimtp/internal/aimtp_controller/store"
	"aimtp/internal/pkg/server"

	"github.com/google/wire"
)

func InitializeServer(*Config) (server.Server, error) {
	wire.Build(
		wire.NewSet(NewWebServer, wire.FieldsOf(new(*Config), "ServerMode")),
		wire.Struct(new(ServerConfig), "*"),
		wire.NewSet(store.ProviderSet, biz.ProviderSet),
		ProvideDB,
		validation.ProviderSet,
	)
	return nil, nil
}