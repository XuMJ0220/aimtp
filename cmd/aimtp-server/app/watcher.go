package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"aimtp/cmd/aimtp-server/app/options"
	"aimtp/internal/aimtp_watcher"
	watcherbiz "aimtp/internal/aimtp_watcher/biz"
	watcherstore "aimtp/internal/aimtp_watcher/store"
	"aimtp/internal/pkg/client"
	"aimtp/pkg/kafka"
	"aimtp/pkg/log"
	"aimtp/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultWatcherInterval   = 120 * time.Second
	defaultWatcherStaleAfter = 10 * time.Minute
	defaultWatcherBatchSize  = 100
)

func NewWatcherCommand(opts *options.ServerOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watcher",
		Short: "AIMTP watcher requeue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatcher(opts)
		},
		Args: cobra.NoArgs,
	}

	return cmd
}

func runWatcher(opts *options.ServerOptions) error {
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
	if cfg.MySQLOptions == nil {
		return fmt.Errorf("mysql options is nil")
	}

	interval := viper.GetDuration("watcher.interval")
	if interval <= 0 {
		interval = defaultWatcherInterval
	}
	staleAfter := viper.GetDuration("watcher.stale-after")
	if staleAfter <= 0 {
		staleAfter = defaultWatcherStaleAfter
	}
	batchSize := viper.GetInt("watcher.batch-size")
	if batchSize <= 0 {
		batchSize = defaultWatcherBatchSize
	}

	controllerClients, err := client.NewControllerClients(cfg.ControllerClusters)
	if err != nil {
		return err
	}
	if len(controllerClients) == 0 {
		log.Infow("Watcher started without controller clients")
	}

	db, err := cfg.MySQLOptions.NewDB()
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	kafkaClient, err := kafka.NewClient(cfg.KafkaOptions)
	if err != nil {
		return err
	}
	producer := kafka.NewProducer(kafkaClient)
	defer producer.Close()

	topic := &kafka.TopicConfig{Topic: cfg.KafkaOptions.Topic}
	store := watcherstore.NewStore(db)
	biz := watcherbiz.New(store, producer, topic, staleAfter, batchSize)
	watcher := aimtp_watcher.New(biz, interval)

	log.Infow("Watcher starting", "interval", interval, "stale_after", staleAfter, "batch_size", batchSize)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return watcher.Run(ctx)
}
