package middleware

import (
	"net/http"
	"strings"

	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

const ContextClaims = "jwt_claims"

const DefaultCookieName = "access_token"

func BearerToken(c *gin.Context, cookieName string) string {
	if cookieName == "" {
		cookieName = DefaultCookieName
	}
	auth := c.GetHeader("Authorization")
	if auth != "" {
		const prefix = "Bearer "
		if strings.HasPrefix(auth, prefix) {
			return strings.TrimSpace(auth[len(prefix):])
		}
	}
	if cookie, err := c.Cookie(cookieName); err == nil {
		return cookie
	}
	return ""
}

func JWTRequired(tokens *jwt.Service, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := BearerToken(c, cookieName)
		claims, err := tokens.Parse(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "برای ادامه وارد حساب خود شوید.",
			})
			return
		}
		c.Set(ContextClaims, claims)
		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) (jwt.Claims, bool) {
	v, ok := c.Get(ContextClaims)
	if !ok {
		return jwt.Claims{}, false
	}
	claims, ok := v.(jwt.Claims)
	return claims, ok
}
