package middleware

import (
	"net/http"

	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

// AdminAccess requires a valid JWT for an admin account.
func AdminAccess(tokens *jwt.Service, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := BearerToken(c, cookieName)
		claims, err := tokens.Parse(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "برای ادامه وارد حساب مدیر شوید.",
			})
			return
		}
		if claims.Kind != jwt.SubjectKindAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "اجازه دسترسی به این بخش را ندارید.",
			})
			return
		}
		c.Set(ContextClaims, claims)
		c.Next()
	}
}

// AdminActorID returns the authenticated admin ID when AdminAccess succeeded.
func AdminActorID(c *gin.Context) (int64, bool) {
	claims, ok := ClaimsFromContext(c)
	if !ok || claims.Kind != jwt.SubjectKindAdmin {
		return 0, false
	}
	id, err := claims.SubjectID()
	if err != nil || !id.Valid() {
		return 0, false
	}
	return int64(id), true
}
