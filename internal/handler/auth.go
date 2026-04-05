package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	token string
}

func NewAuthHandler(token string) *AuthHandler {
	return &AuthHandler{token: token}
}

func (h *AuthHandler) LoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", nil)
}

func (h *AuthHandler) Login(c echo.Context) error {
	if c.FormValue("token") != h.token {
		return c.Render(http.StatusUnauthorized, "login.html", map[string]any{
			"Error": "Invalid token.",
		})
	}
	c.SetCookie(&http.Cookie{
		Name:     "auth_token",
		Value:    h.token,
		Path:     "/",
		Expires:  time.Now().Add(90 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return c.Redirect(http.StatusSeeOther, "/")
}
