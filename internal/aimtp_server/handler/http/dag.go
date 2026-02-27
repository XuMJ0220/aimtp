package http

import (
	"aimtp/pkg/core"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateDAG(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.DAGV1().CreateDAG,h.val.ValidateCreateDAGRequest)
}
