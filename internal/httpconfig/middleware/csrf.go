package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const CSRFCookieName = "csrftoken"
const CSRFHeaderName = "X-CSRFToken"

// EnsureCSRF sets a non-HttpOnly CSRF cookie when missing (double-submit pattern).
func EnsureCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := c.Cookie(CSRFCookieName); err != nil {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err == nil {
				token := hex.EncodeToString(b)
				c.SetSameSite(http.SameSiteLaxMode)
				c.SetCookie(CSRFCookieName, token, 86400*7, "/", "", false, false)
			}
		}
		c.Next()
	}
}

// CSRFRequired rejects mutating requests when the header does not match the cookie.
func CSRFRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		cookie, err := c.Cookie(CSRFCookieName)
		if err != nil || cookie == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "برای این عملیات نشست امن معتبر لازم است. صفحه را یک‌بار تازه کنید.",
			})
			return
		}
		header := strings.TrimSpace(c.GetHeader(CSRFHeaderName))
		if header == "" || header != cookie {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "درخواست امنیتی تأیید نشد. صفحه را تازه کنید و دوباره تلاش کنید.",
			})
			return
		}
		c.Next()
	}
}
