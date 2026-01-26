package middlewares

import (
	"net/http"
	"strings"

	"github.com/Shivamrai15/rs-safari/models"
	"github.com/Shivamrai15/rs-safari/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header missing",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		jwtSecret := []byte(utils.GetConfigValue("JWT_ACCESS_SECRET"))
		
		token, err := jwt.ParseWithClaims(tokenString, &models.TokenPayload{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token",
			})
			c.Abort()
			return
		}
		
		c.Set("user", token.Claims.(*models.TokenPayload))
		c.Next()
	}
}
