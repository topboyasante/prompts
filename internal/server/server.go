package server

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const requestIDKey = "request_id"
const loggerKey = "logger"

func New() *gin.Engine {
	r := gin.New()
	r.MaxMultipartMemory = 10 << 20

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	r.Use(requestIDMiddleware())
	r.Use(recoveryMiddleware())
	r.Use(loggerMiddleware(logger))
	r.Use(corsMiddleware())

	return r
}

func RequestIDFromContext(c *gin.Context) string {
	v, ok := c.Get(requestIDKey)
	if !ok {
		return ""
	}
	id, _ := v.(string)
	return id
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.NewString()
		c.Set(requestIDKey, id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Set(loggerKey, logrus.WithField("request_id", id))
		c.Next()
	}
}

func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				LoggerFromContext(c).WithField("panic", rec).Error("panic recovered")
				RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}

func loggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		LoggerFromContext(c).WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"query":  c.Request.URL.RawQuery,
		}).Info("http request started")
		c.Next()

		entry := logger.WithFields(logrus.Fields{
			"request_id":     RequestIDFromContext(c),
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"query":          c.Request.URL.RawQuery,
			"status":         c.Writer.Status(),
			"latency_ms":     time.Since(start).Milliseconds(),
			"user_id":        c.GetString("user_id"),
			"remote_ip":      c.ClientIP(),
			"user_agent":     c.Request.UserAgent(),
			"response_bytes": c.Writer.Size(),
			"errors":         c.Errors.String(),
		})
		if c.Writer.Status() >= http.StatusInternalServerError {
			entry.Error("http request completed")
			return
		}
		if c.Writer.Status() >= http.StatusBadRequest {
			entry.Warn("http request completed")
			return
		}
		entry.Info("http request completed")
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func LoggerFromContext(c *gin.Context) *logrus.Entry {
	v, ok := c.Get(loggerKey)
	if ok {
		if entry, ok := v.(*logrus.Entry); ok {
			return entry
		}
	}
	return logrus.NewEntry(logrus.StandardLogger()).WithField("request_id", RequestIDFromContext(c))
}
