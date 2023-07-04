package middleware

import (
	"aliagha/helpers"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			authHeader := ctx.Request().Header.Get("Authorization")
			if authHeader == "" {
				return ctx.NoContent(http.StatusUnauthorized)
			}

			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
			userID, err := helpers.ParseJWTToken(tokenString, secretKey)

			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, "Invalid token")
			}

			ctx.Set("user_id", userID)
			return next(ctx)
		}
	}
}
