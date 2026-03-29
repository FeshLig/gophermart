package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[%d] %s %s (%s)\n", status, c.Request.Method, c.Request.URL.Path, latency)
	}
}
