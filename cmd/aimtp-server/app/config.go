// Copyright 2024 许铭杰 (1044011439@qq.com). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"aimtp/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// defaultHomeDir 定义放置 aimtp 服务配置的默认目录.
	defaultHomeDir = ".aimtp"
	// defaultConfigName 指定 aimtp 服务的默认配置文件名.
	defaultConfigName = "aimtp-server.yaml"
)

// onInitialize 设置需要读取的配置文件名、环境变量，并将其内容读取到 viper 中.
func onInitialize() {
	// 如果请求了版本标志，则跳过配置初始化
	if version.IsVersionRequested() {
		return
	}

	if configFile != "" {
		// 从命令行选项指定的配置文件中读取
		viper.SetConfigFile(configFile)
	} else {
		// 使用默认配置文件路径和名称
		for _, dir := range searchDirs() {
			// 将 dir 目录加入到配置文件的搜索路径
			viper.AddConfigPath(dir)
		}

		// 设置配置文件格式为 YAML
		viper.SetConfigType("yaml")

		// 配置文件名称（没有文件扩展名）
		viper.SetConfigName(defaultConfigName)
	}

	// 读取环境变量并设置前缀
	setupEnvironmentVariables()

	// 读取配置文件, 如果指定了配置文件名, 则使用指定的配置文件, 否则在注册的搜索路径中搜索
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Failed to read viper configuration file, err: %v", err)
	}

	// 打印当前使用的配置文件，方便调试
	log.Printf("Using config file: %s", viper.ConfigFileUsed())
}

// setupEnvironmentVariables 配置环境变量规则.
func setupEnvironmentVariables() {
	// viper 开启读取环境变量的规则
	viper.AutomaticEnv()
	// 设置环境变量前缀, 前缀可以防止与其他同名环境造成冲突
	viper.SetEnvPrefix("MINIBLOG")
	// 替换环境变量 key 中的分割副 '.' 和 '-' 为 '_'
	replace := strings.NewReplacer(".", "_", "-", "_")
	viper.SetEnvKeyReplacer(replace)
}

// searchDirs 返回默认的配置文件搜索目录.
func searchDirs() []string {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	return []string{filepath.Join(homeDir, defaultHomeDir), "."}
}

// filePath 获取默认配置文件的完整路径.
func filePath() string {
	home, err := os.UserHomeDir()
	// 如果不能获取用户主目录，则记录错误并返回空路径
	cobra.CheckErr(err)
	return filepath.Join(home, defaultHomeDir, defaultConfigName)
}