package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// Check godoc
// @Summary 健康检查
// @Description 检查系统健康状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "健康状态"
// @Router /api/health [get]
func (h *HealthController) Check(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"status": "healthy",
		},
	})
}
