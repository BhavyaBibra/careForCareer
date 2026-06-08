package middleware

import (
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const UserIDKey = "user_id"

// JWT returns a Gin middleware that validates RS256 JWT access tokens.
func JWT(publicKey *rsa.PublicKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Missing or invalid Authorization header"))
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return publicKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Invalid or expired token"))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Invalid claims"))
			return
		}

		sub, _ := claims["sub"].(string)
		userID, err := uuid.Parse(sub)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Invalid user ID in token"))
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from the Gin context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func errorResponse(code, message string) gin.H {
	return gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
}
