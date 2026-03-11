package auth

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/topboyasante/prompts/internal/server"
)

type userIDContextKey struct{}

const userIDGinKey = "user_id"

type Middleware struct {
	jwtSecret string
}

func NewMiddleware(jwtSecret string) *Middleware {
	return &Middleware{jwtSecret: jwtSecret}
}

func (m *Middleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var rawToken string
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(header, "Bearer ") {
			rawToken = strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		} else {
			cookie, err := c.Cookie("prompts_token")
			if err != nil || strings.TrimSpace(cookie) == "" {
				server.RespondError(c, 401, "UNAUTHORIZED", "missing bearer token")
				c.Abort()
				return
			}
			rawToken = strings.TrimSpace(cookie)
		}

		claims, err := ParseToken(rawToken, m.jwtSecret)
		if err != nil {
			server.RespondError(c, 401, "UNAUTHORIZED", "invalid access token")
			c.Abort()
			return
		}

		ctx := context.WithValue(c.Request.Context(), userIDContextKey{}, claims.UserID)
		c.Request = c.Request.WithContext(ctx)
		c.Set(userIDGinKey, claims.UserID)
		c.Next()
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDContextKey{}).(string)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}

func UserIDFromGin(c *gin.Context) (string, bool) {
	v, ok := c.Get(userIDGinKey)
	if !ok {
		return "", false
	}
	userID, ok := v.(string)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}
