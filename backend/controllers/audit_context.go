package controllers

import (
	"mcloud/services"

	"github.com/gin-gonic/gin"
)

// auditContext 从 JWT 上下文取操作人快照 + 生成本请求聚合 requestID。
// user_id 缺失时返回零值 Operator（理论上受保护路由不会发生）。
func auditContext(c *gin.Context) (services.Operator, string) {
	requestID := services.NewRequestID()
	userID, exists := c.Get("user_id")
	if !exists {
		return services.Operator{}, requestID
	}
	uid, ok := userID.(uint)
	if !ok {
		return services.Operator{}, requestID
	}
	op, err := services.GetOperator(uid)
	if err != nil {
		return services.Operator{ID: uid}, requestID
	}
	return op, requestID
}
