// Package middleware provides HTTP middleware functions for rate limiting, metrics, logging, and error handling.
package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"url-shorterner/internal/log"

	"github.com/gin-gonic/gin"
)

// Recovery returns a Gin middleware that recovers from panics and logs them with stack traces.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		stackTrace := getPanicStackTrace()
		log.Error("Panic recovered: %v | %s %s | %s\n%s",
			recovered,
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			stackTrace,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		c.Abort()
	})
}

// getPanicStackTrace returns a formatted stack trace for panic recovery.
func getPanicStackTrace() string {
	stack := debug.Stack()
	lines := bytes.Split(stack, []byte("\n"))

	var filtered []string
	for i, line := range lines {
		lineStr := string(line)

		// Skip goroutine header
		if strings.HasPrefix(lineStr, "goroutine") {
			continue
		}

		// Skip runtime and framework internal frames
		if strings.Contains(lineStr, "runtime/") ||
			strings.Contains(lineStr, "gin-gonic/gin") ||
			strings.Contains(lineStr, "internal/middleware") ||
			strings.Contains(lineStr, "internal/log") {
			continue
		}

		// Include frames from our application code
		if strings.Contains(lineStr, "url-shorterner/") {
			// Include this line and the next (file:line)
			if i+1 < len(lines) {
				nextLine := string(lines[i+1])
				if strings.HasPrefix(nextLine, "\t") {
					filtered = append(filtered, lineStr)
					filtered = append(filtered, nextLine)
				}
			} else {
				filtered = append(filtered, lineStr)
			}
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	return fmt.Sprintf("Stack trace:\n%s", strings.Join(filtered, "\n"))
}
