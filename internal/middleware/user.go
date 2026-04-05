package middleware

import (
	"net/http"
	"strconv"

	"github.com/ivansantos/rustypushups/internal/store"
	"github.com/labstack/echo/v4"
)

const UserIDCookie = "user_id"
const UserContextKey = "current_user_id"

func UserRequired(s *store.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(UserIDCookie)
			if err != nil || cookie.Value == "" {
				return c.Redirect(http.StatusSeeOther, "/users")
			}
			userID, err := strconv.ParseInt(cookie.Value, 10, 64)
			if err != nil {
				return c.Redirect(http.StatusSeeOther, "/users")
			}
			user, err := s.GetUser(c.Request().Context(), userID)
			if err != nil {
				return c.Redirect(http.StatusSeeOther, "/users")
			}
			c.Set(UserContextKey, user)
			return next(c)
		}
	}
}
