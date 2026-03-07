//go:build wireinject
// +build wireinject

package aimtp_server

import (
	"aimtp/internal/aimtp_server/biz"
	"aimtp/internal/aimtp_server/pkg/validation"
	"aimtp/internal/aimtp_server/store"
	"aimtp/internal/pkg/client"
	"aimtp/internal/pkg/server"
	"aimtp/pkg/kafka"

	"github.com/google/wire"
)

func InitializeServer(*Config) (server.Server, error) {
	wire.Build(
		wire.NewSet(NewWebServer, wire.FieldsOf(new(*Config), "ServerMode")),
		wire.Struct(new(ServerConfig), "*"),
		wire.NewSet(store.ProviderSet, biz.ProviderSet),
		ProvideDB,
		validation.ProviderSet,
		wire.FieldsOf(new(*Config), "ControllerClusters"),
		client.NewControllerClients,
		wire.FieldsOf(new(*Config), "KafkaOptions"),
		kafka.ProvideClient,
		kafka.ProvideProducer,
		kafka.ProvideTopicConfig,
	)
	return nil, nil
}
