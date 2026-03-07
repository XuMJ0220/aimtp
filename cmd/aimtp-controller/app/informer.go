package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"aimtp/cmd/aimtp-controller/app/options"
	"aimtp/internal/aimtp_controller/informer"
	"aimtp/internal/aimtp_controller/store"
	"aimtp/internal/pkg/k8s"
	"aimtp/pkg/log"
	"aimtp/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewInformerCommand(opts *options.ServerOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "informer",
		Short: "Start AIMTP Informer Controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInformer(opts)
		},
		Args: cobra.NoArgs,
	}

	return cmd
}

func runInformer(opts *options.ServerOptions) error {
	version.PrintAndExitIfRequested()

	log.Init(logOptions())
	defer log.Sync()

	// 1. 加载配置
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

	if opts.InformerCluster == "" {
		return fmt.Errorf("informer-cluster is required")
	}

	// 2. 初始化 MySQL Store
	if cfg.MySQLOptions == nil {
		return fmt.Errorf("mysql options is nil")
	}
	db, err := cfg.MySQLOptions.NewDB()
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// 初始化全局 Store
	store.NewStore(db)

	// 3. 初始化 K8s Config
	restCfg, err := k8s.NewRestConfig(*cfg.K8sOptions)
	if err != nil {
		return fmt.Errorf("create k8s rest config: %w", err)
	}

	// 4. 创建 Informer Controller
	// 注意：store.S.Job() 是全局 Store 的 JobStore 实现
	controller, err := informer.New(
		opts.InformerCluster,
		restCfg,
		store.S.Job(),
		store.S.Pod(),
	)
	if err != nil {
		return err
	}

	log.Infow("Informer Controller starting", "cluster", opts.InformerCluster)

	// 5. 启动
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return controller.Run(ctx.Done())
}
