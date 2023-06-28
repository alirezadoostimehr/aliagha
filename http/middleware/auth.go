package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func AuthenticatorMiddleware(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, "Missing Authorization header")
			}

			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("invalid token signing method")
				}
				return []byte(secretKey), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, "Invalid or expired token")
			}

			claims := token.Claims.(jwt.MapClaims)
			userID := claims["user_id"].(string)
			c.Set("user_id", userID)

			return next(c)
		}
	}
}
