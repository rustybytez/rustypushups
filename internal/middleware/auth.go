package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const authCookie = "auth_token"

func TokenRequired(token string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(authCookie)
			if err != nil || cookie.Value != token {
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			return next(c)
		}
	}
}
