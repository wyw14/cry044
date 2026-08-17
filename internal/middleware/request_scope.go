package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var reviewResponseHeaders = http.Header{
	"X-Content-Type-Options":       {"nosniff"},
	"X-Frame-Options":              {"DENY"},
	"Referrer-Policy":              {"same-origin"},
	"Access-Control-Allow-Origin":  {"http://localhost:5173"},
	"Access-Control-Allow-Headers": {"Content-Type,X-Actor-ID,X-Actor-Role,If-Match,Idempotency-Key"},
	"Access-Control-Allow-Methods": {"GET,POST,OPTIONS"},
}

func ReviewBoundary(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := traceToken()
		c.Set("request_id", requestID)
		headers := c.Writer.Header()
		headers.Set("X-Request-ID", requestID)
		for name, values := range reviewResponseHeaders {
			headers[name] = append([]string(nil), values...)
		}
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		limited, cancel := context.WithDeadline(c.Request.Context(), time.Now().Add(timeout))
		defer cancel()
		c.Request = c.Request.Clone(limited)
		c.Next()
	}
}

func traceToken() string {
	buffer := make([]byte, 12)
	if count, err := rand.Read(buffer); err != nil || count != len(buffer) {
		return "review-local"
	}
	return "rv-" + base64.RawURLEncoding.EncodeToString(buffer)
}
