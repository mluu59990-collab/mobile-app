package middleware

import (
	"net/http"
	"strings"

	"expense-app/backend/internal/apierror"
	"expense-app/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "auth_user_id"

func JWT(tokens auth.Tokens) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			apierror.Write(c, http.StatusUnauthorized, "invalid_access_token", "A valid bearer access token is required")
			return
		}
		id, err := tokens.Parse(parts[1])
		if err != nil {
			apierror.Write(c, http.StatusUnauthorized, "invalid_access_token", "A valid bearer access token is required")
			return
		}
		c.Set(UserIDKey, id)
		c.Next()
	}
}
