// Copyright 2024 许铭杰 (1044011439@qq.com). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	"aimtp/internal/aimtp_server"
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
	aimtp_server.GinServerMode,
	aimtp_server.GRPCServerMode,
	aimtp_server.GRPCGatewayServerMode,
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
	TLSOptions         *genericoptions.TLSOptions `json:"tls" mapstructure:"tls"`
	ControllerClusters map[string]string          `json:"controller-clusters" mapstructure:"controller-clusters"`
}

// NewServerOptions 创建带有默认值的 ServerOptions 实例.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
		ServerMode:   aimtp_server.GRPCGatewayServerMode,
		JWTKey:       "Rtg8BPKNEf2mB4mgvKONGPZZQSaJWNLijxR42qRgq0iBb5",
		Expiration:   2 * time.Hour,
		GRPCOptions:  genericoptions.NewGRPCOptions(),
		HTTPOptions:  genericoptions.NewHTTPOptions(),
		MySQLOptions: genericoptions.NewMySQLOptions(),
		TLSOptions:   genericoptions.NewTLSOptions(),
	}
	opts.GRPCOptions.Addr = ":6667"
	opts.HTTPOptions.Addr = ":5556"
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
	if stringsutil.StringIn(o.ServerMode, []string{aimtp_server.GRPCServerMode, aimtp_server.GRPCGatewayServerMode}) {
		errs = append(errs, o.GRPCOptions.Validate()...)
	}

	// 合并所有错误并返回
	return utilerrors.NewAggregate(errs)
}

// Config 基于 ServerOptions 构建运行时配置 apiserver.Config.
// ----------- 在运行时配置可用 -----------
func (o *ServerOptions) Config() (*aimtp_server.Config, error) {
	return &aimtp_server.Config{
		ServerMode:         o.ServerMode,
		JWTKey:             o.JWTKey,
		Expiration:         o.Expiration,
		GRPCOptions:        o.GRPCOptions,
		HTTPOptions:        o.HTTPOptions,
		MySQLOptions:       o.MySQLOptions,
		TLSOptions:         o.TLSOptions,
		ControllerClusters: o.ControllerClusters,
	}, nil
}
