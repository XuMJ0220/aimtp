// Copyright 2024 许铭杰 (1044011439@qq.com). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package log

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	// 确保日志缓冲区被刷新
	Sync()

	// 自定义配置文件
	opts := &Options{
		Level:             "debug",            // 设置日志级别为 debug
		Format:            "json",             // 设置日志格式为 JSON
		DisableCaller:     false,              // 显示调用日志的文件和行号
		DisableStacktrace: false,              // 允许打印堆栈信息
		OutputPaths:       []string{"stdout"}, // 将日志输出到标准输出
	}

	// 初始化全局日志对象
	Init(opts)

	// 测试不同级别的日志输出
	Debugw("This is a debug message", "key1", "value1", "key2", 123)
	Infow("This is an info message", "key", "value")
	Warnw("This is a warning message", "timestamp", time.Now())
	Errorw("This is an error message", "error", "something went wrong")
}
