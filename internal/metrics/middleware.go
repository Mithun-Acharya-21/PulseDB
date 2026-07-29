package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GinPrometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // process the request

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath() 
		if path == "" {
			path = "unknown"
		}

		elapsed := time.Since(start).Seconds()

		HTTPRequestTotal.WithLabelValues(method, path, status).Inc()
		HttpRequestDuration.WithLabelValues(method, path).Observe(elapsed)
	}
}
