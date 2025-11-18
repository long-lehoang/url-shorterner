// Package middleware provides HTTP middleware functions for rate limiting, metrics, logging, and error handling.
package middleware

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"url-shorterner/internal/log"

	"github.com/gin-gonic/gin"
)

// Logger returns a Gin middleware that logs HTTP requests and responses.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errors := c.Errors.ByType(gin.ErrorTypePrivate)

		if raw != "" {
			path = path + "?" + raw
		}

		if len(errors) > 0 {
			var errorMessages []string
			for _, err := range errors {
				errorMsg := err.Error()
				stackTrace := getStackTrace()
				if stackTrace != "" {
					errorMsg = fmt.Sprintf("%s\n%s", errorMsg, stackTrace)
				}
				errorMessages = append(errorMessages, errorMsg)
			}
			errorMessage := strings.Join(errorMessages, " | ")
			log.Error("[%s] %s %s %d %v | %s | %s",
				clientIP,
				method,
				path,
				statusCode,
				latency,
				errorMessage,
				c.Request.UserAgent(),
			)
		} else {
			log.Info("[%s] %s %s %d %v | %s",
				clientIP,
				method,
				path,
				statusCode,
				latency,
				c.Request.UserAgent(),
			)
		}
	}
}

// getStackTrace returns a formatted stack trace filtered to show only application code.
func getStackTrace() string {
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
			strings.Contains(lineStr, "internal/log") ||
			strings.Contains(lineStr, "net/http") {
			continue
		}

		// Include frames from our application code
		if strings.Contains(lineStr, "url-shorterner/") {
			// Skip getStackTrace function itself
			if strings.Contains(lineStr, "getStackTrace") {
				continue
			}
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
