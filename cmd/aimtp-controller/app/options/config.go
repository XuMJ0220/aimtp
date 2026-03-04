// Copyright 2024 许铭杰 (1044011439@qq.com). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	"aimtp/internal/aimtp_controller"
	"aimtp/internal/pkg/k8s"
	"errors"
	"fmt"
	"time"

	genericoptions "aimtp/pkg/options"
	stringsutil "aimtp/pkg/util/strings"

	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"k8s.io/apimachinery/pkg/util/sets"
)

// 定义支持的服务器模式集合
var availableServerModes = sets.New(
	aimtp_controller.GinServerMode,
	aimtp_controller.GRPCServerMode,
	aimtp_controller.GRPCGatewayServerMode,
)

// ServerOptions 包含服务器配置选项.
type ServerOptions struct {
	// ServerMode 定义服务器模式：gRPC、Gin HTTP、HTTP Reverse Proxy.
	ServerMode string `json:"server-mode" mapstructure:"server-mode"`
	// JWTKey 定义 JWT 密钥.
	JWTKey string `json:"jwt-key" mapstructure:"jwt-key"`
	// Expiration 定义 JWT Token 的过期时间.
	Expiration time.Duration `json:"expiration" mapstructure:"expiration"`
	// GRPCOptions 包含 gRPC 配置选项.
	GRPCOptions *genericoptions.GRPCOptions `json:"grpc" mapstructure:"grpc"`
	// HTTPOptions 包含 HTTP 配置选项.
	HTTPOptions *genericoptions.HTTPOptions `json:"http" mapstructure:"http"`
	// MySQLOptions 包含 MySQL 配置选项.
	MySQLOptions *genericoptions.MySQLOptions `json:"mysql" mapstructure:"mysql"`
	// TLSOptions 包含 TLS 配置选项.
	TLSOptions *genericoptions.TLSOptions `json:"tls" mapstructure:"tls"`
	// KafkaOptions 包含 Kafka 配置选项.
	KafkaOptions *genericoptions.KafkaOptions `json:"kafka" mapstructure:"kafka"`
	K8sOptions   *k8s.ConfigOptions           `json:"k8s" mapstructure:"k8s"`
}

// NewServerOptions 创建带有默认值的 ServerOptions 实例.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
		ServerMode:   aimtp_controller.GRPCGatewayServerMode,
		JWTKey:       "Rtg8BPKNEf2mB4mgvKONGPZZQSaJWNLijxR42qRgq0iBb5",
		Expiration:   2 * time.Hour,
		GRPCOptions:  genericoptions.NewGRPCOptions(),
		HTTPOptions:  genericoptions.NewHTTPOptions(),
		MySQLOptions: genericoptions.NewMySQLOptions(),
		TLSOptions:   genericoptions.NewTLSOptions(),
		KafkaOptions: genericoptions.NewKafkaOptions(),
		K8sOptions:   &k8s.ConfigOptions{},
	}
	opts.GRPCOptions.Addr = ":6668"
	opts.HTTPOptions.Addr = ":5557"
	opts.MySQLOptions.Addr = ":3306"
	return opts
}

// AddFlags 将 ServerOptions 的选项绑定到命令行标志.
// 通过使用 pflag 包，可以实现从命令行中解析这些选项的功能.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// 命令行 --server-mode 将绑定到 o.ServerMode, 若命令行不包含 --server-mode, 则用默认值 o.ServerMode
	// 例如 --server-mode=gin
	fs.StringVar(&o.ServerMode, "server-mode", o.ServerMode, fmt.Sprintf("Server mode, available options: %v", availableServerModes.UnsortedList()))
	fs.StringVar(&o.JWTKey, "jwt-key", o.JWTKey, "JWT signing key. Must be at least 6 characters long.")
	// 命令行 --expiration 将绑定到 o.Expiration, 若命令行不包含 --server-mode, 则用默认值 o.Expiration
	// 例如 --expiration=4h
	fs.DurationVar(&o.Expiration, "expiration", o.Expiration, "The expiration duration of JWT tokens.")

	o.GRPCOptions.AddFlags(fs, "grpc")
	o.HTTPOptions.AddFlags(fs, "http")
	o.MySQLOptions.AddFlags(fs, "mysql")
	o.TLSOptions.AddFlags(fs, "tls")
	o.KafkaOptions.AddFlags(fs, "kafka")
	fs.StringVar(&o.K8sOptions.Kubeconfig, "k8s.kubeconfig", o.K8sOptions.Kubeconfig, "")
	fs.BoolVar(&o.K8sOptions.InCluster, "k8s.in-cluster", o.K8sOptions.InCluster, "")
	fs.Float32Var(&o.K8sOptions.QPS, "k8s.qps", o.K8sOptions.QPS, "")
	fs.IntVar(&o.K8sOptions.Burst, "k8s.burst", o.K8sOptions.Burst, "")
	fs.DurationVar(&o.K8sOptions.Timeout, "k8s.timeout", o.K8sOptions.Timeout, "")
	fs.StringVar(&o.K8sOptions.UserAgent, "k8s.user-agent", o.K8sOptions.UserAgent, "")
	fs.BoolVar(&o.K8sOptions.Insecure, "k8s.insecure", o.K8sOptions.Insecure, "Skip TLS verification for k8s client")
}

// Validate 校验 ServerOptions 中的选项是否合法.
func (o *ServerOptions) Validate() error {
	errs := []error{}

	// 校验 ServerMode 是否有效
	if !availableServerModes.Has(o.ServerMode) {
		errs = append(errs, fmt.Errorf("invalid server mode: must be one of %v", availableServerModes.UnsortedList()))
	}

	// 校验 JWTKey 长度
	if len(o.JWTKey) < 6 {
		errs = append(errs, errors.New("JWTKey must be at least 6 characters long"))
	}

	// 如果是 gRPC 或 gRPC-Gateway 模式, 校验 gRPC 配置
	if stringsutil.StringIn(o.ServerMode, []string{aimtp_controller.GRPCServerMode, aimtp_controller.GRPCGatewayServerMode}) {
		errs = append(errs, o.GRPCOptions.Validate()...)
	}

	// 如果是 HTTP 模式，校验 HTTP 配置
	if stringsutil.StringIn(o.ServerMode, []string{aimtp_controller.GinServerMode}) {
		errs = append(errs, o.HTTPOptions.Validate()...)
	}

	// 校验 MySQL 配置
	errs = append(errs, o.MySQLOptions.Validate()...)

	// 校验 TLS 配置
	errs = append(errs, o.TLSOptions.Validate()...)

	// 校验 kafka 配置
	errs = append(errs, o.KafkaOptions.Validate()...)

	// 合并所有错误并返回
	return utilerrors.NewAggregate(errs)
}

// Config 基于 ServerOptions 构建运行时配置 apiserver.Config.
// ----------- 在运行时配置可用 -----------
func (o *ServerOptions) Config() (*aimtp_controller.Config, error) {
	return &aimtp_controller.Config{
		ServerMode:   o.ServerMode,
		JWTKey:       o.JWTKey,
		Expiration:   o.Expiration,
		GRPCOptions:  o.GRPCOptions,
		HTTPOptions:  o.HTTPOptions,
		MySQLOptions: o.MySQLOptions,
		TLSOptions:   o.TLSOptions,
		KafkaOptions: o.KafkaOptions,
		K8sOptions:   o.K8sOptions,
	}, nil
}
