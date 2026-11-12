package middleware

import (
	"net/http"

	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

// PanelAccess requires a valid JWT for a teacher account (school owners are teachers).
func PanelAccess(tokens *jwt.Service, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := BearerToken(c, cookieName)
		claims, err := tokens.Parse(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "برای ادامه وارد حساب خود شوید.",
			})
			return
		}
		if !panelRoleAllowed(claims) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "اجازه دسترسی به این بخش را ندارید.",
			})
			return
		}
		c.Set(ContextClaims, claims)
		c.Next()
	}
}

func panelRoleAllowed(claims jwt.Claims) bool {
	return claims.Kind == jwt.SubjectKindUser && claims.Role == "teacher"
}

// TeacherUserID returns the authenticated teacher ID when PanelAccess succeeded.
func TeacherUserID(c *gin.Context) (int64, bool) {
	claims, ok := ClaimsFromContext(c)
	if !ok || claims.Kind != jwt.SubjectKindUser {
		return 0, false
	}
	id, err := claims.SubjectID()
	if err != nil || !id.Valid() {
		return 0, false
	}
	return int64(id), true
}
