package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid authorization header"})
			}

			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			}

			userID, ok := claims["user_id"].(float64)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user_id in token"})
			}

			c.Set("user_id", uint(userID))
			return next(c)
		}
	}
}
