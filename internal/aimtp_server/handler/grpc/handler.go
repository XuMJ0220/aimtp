// Copyright 2024 许铭杰 (1044011439@qq.com). All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package grpc

import (
	apiv1 "aimtp/pkg/api/aimtp_server/v1"
)

// Handler 负责处理模块的请求.
type Handler struct {
	apiv1.UnimplementedAIMTP_SERVERServer // 提供默认实现

}

func NewHandler() *Handler {
	return &Handler{}
}
