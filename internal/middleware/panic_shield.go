package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func RecoverReviewAPI(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			panicValue := recover()
			if panicValue == nil {
				return
			}
			requestID := c.GetString("request_id")
			logger.Error("review room recovered from panic", zap.String("request_id", requestID), zap.String("panic_type", fmt.Sprintf("%T", panicValue)))
			c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{
				"code": "UNEXPECTED_REVIEW_FAILURE", "message": "internal review workflow error", "field_errors": map[string]string{}, "request_id": requestID,
			})
		}()
		c.Next()
	}
}
