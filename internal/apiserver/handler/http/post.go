package http

import (
	"aimtp/pkg/core"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreatePost(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.PostV1().Create, h.val.ValidateCreatePostRequest)
}

func (h *Handler) UpdatePost(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.PostV1().Update, h.val.ValidateUpdatePostRequest)
}

func (h *Handler) DeletePost(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.PostV1().Delete, h.val.ValidateDeletePostRequest)
}

func (h *Handler) GetPost(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.PostV1().Get, h.val.ValidateGetPostRequest)
}

func (h *Handler) ListPost(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.PostV1().List, h.val.ValidateListPostRequest)
}
