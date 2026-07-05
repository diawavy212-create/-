package middleware

import (
	"net/http"
	"strings"

	"teacher-platform/server/internal/config"
	"teacher-platform/server/internal/response"
	"teacher-platform/server/internal/security"

	"github.com/gin-gonic/gin"
)

func RequireToken(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			response.Fail(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		claims, err := security.VerifyToken(cfg.AuthTokenSecret, token)
		if err != nil {
			if cfg.DevAuthEnabled {
				claims = devClaims(token)
				if claims.UserID == 0 {
					response.Fail(c, http.StatusUnauthorized, "invalid bearer token")
					c.Abort()
					return
				}
			} else {
				response.Fail(c, http.StatusUnauthorized, "invalid bearer token")
				c.Abort()
				return
			}
		}

		c.Set("subject", token)
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func devClaims(token string) security.Claims {
	switch token {
	case "dev-mini-token":
		return security.Claims{UserID: 1, Role: "teacher"}
	case "dev-admin-token":
		return security.Claims{UserID: 2, Role: "party_admin"}
	default:
		return security.Claims{}
	}
}
