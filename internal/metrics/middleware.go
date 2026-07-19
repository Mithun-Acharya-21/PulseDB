package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GinPrometheus returns a Gin middleware that records HTTP request
// count (HTTPRequestTotal) and latency (HttpRequestDuration) for
// every request processed by the router.
func GinPrometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // process the request

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath() // template path like /api/v1/monitors/:id
		if path == "" {
			path = "unknown"
		}

		elapsed := time.Since(start).Seconds()

		HTTPRequestTotal.WithLabelValues(method, path, status).Inc()
		HttpRequestDuration.WithLabelValues(method, path).Observe(elapsed)
	}
}
