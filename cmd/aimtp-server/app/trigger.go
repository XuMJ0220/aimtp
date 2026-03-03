package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"aimtp/cmd/aimtp-server/app/options"
	"aimtp/internal/aimtp_trigger"
	triggerbiz "aimtp/internal/aimtp_trigger/biz"
	"aimtp/internal/pkg/client"
	"aimtp/pkg/kafka"
	"aimtp/pkg/log"
	"aimtp/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTriggerCommand(opts *options.ServerOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "AIMTP trigger consumer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrigger(opts)
		},
		Args: cobra.NoArgs,
	}

	return cmd
}

func runTrigger(opts *options.ServerOptions) error {
	version.PrintAndExitIfRequested()

	log.Init(logOptions())
	defer log.Sync()

	if err := viper.Unmarshal(opts); err != nil {
		return err
	}
	if err := opts.Validate(); err != nil {
		return err
	}

	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	if cfg.KafkaOptions == nil {
		return fmt.Errorf("kafka options is nil")
	}
	if cfg.KafkaOptions.Topic == "" {
		return fmt.Errorf("kafka topic is empty")
	}
	if cfg.KafkaOptions.ReaderOptions.GroupID == "" {
		cfg.KafkaOptions.ReaderOptions.GroupID = "aimtp-trigger"
	}

	log.Infow("Trigger consumer starting", "topic", cfg.KafkaOptions.Topic, "group_id", cfg.KafkaOptions.ReaderOptions.GroupID, "brokers", cfg.KafkaOptions.Brokers)

	controllerClients, err := client.NewControllerClients(cfg.ControllerClusters)
	if err != nil {
		return err
	}
	kafkaClient, err := kafka.NewClient(cfg.KafkaOptions)
	if err != nil {
		return err
	}

	consumer := kafka.NewConsumer(kafkaClient)
	defer consumer.Close()
	biz := triggerbiz.New(controllerClients)
	trigger := aimtp_trigger.New(consumer, biz)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- trigger.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		if err == context.Canceled {
			return nil
		}
		return err
	}
}
