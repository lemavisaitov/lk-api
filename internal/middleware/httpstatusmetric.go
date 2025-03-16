package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/lemavisaitov/lk-api/internal/logger"
	"github.com/lemavisaitov/lk-api/internal/metrics"
	"go.uber.org/zap"
)

func HttpStatusMetric() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		statusCode := c.Writer.Status()
		method := c.Request.Method

		logger.Info("HTTP request processed",
			zap.Int("status_code", statusCode),
			zap.String("method", method),
		)

		metrics.HttpStatusMetricInc(statusCode, method)
	}
}
