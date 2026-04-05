package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ivansantos/rustypushups/internal/db"
	appmw "github.com/ivansantos/rustypushups/internal/middleware"
	"github.com/ivansantos/rustypushups/internal/store"
	"github.com/labstack/echo/v4"
)

type UsersHandler struct {
	store *store.Store
}

func NewUsersHandler(s *store.Store) *UsersHandler {
	return &UsersHandler{store: s}
}

func (h *UsersHandler) List(c echo.Context) error {
	users, err := h.store.ListUsers(c.Request().Context())
	if err != nil {
		return err
	}

	data := map[string]any{
		"Users": users,
	}

	// If a user is already selected, load their current month goal
	if cookie, err := c.Cookie(appmw.UserIDCookie); err == nil && cookie.Value != "" {
		if userID, err := strconv.ParseInt(cookie.Value, 10, 64); err == nil {
			if user, err := h.store.GetUser(c.Request().Context(), userID); err == nil {
				now := time.Now()
				data["User"] = user
				goal, err := h.store.GetMonthlyGoal(c.Request().Context(), db.GetMonthlyGoalParams{
					Year:  int64(now.Year()),
					Month: int64(now.Month()),
				})
				if err == nil {
					data["MonthGoal"] = goal.Goal
					data["GoalSet"] = true
				} else if !errors.Is(err, sql.ErrNoRows) {
					return err
				}
				data["MonthLabel"] = now.Format("January 2006")
			}
		}
	}

	return c.Render(http.StatusOK, "users.html", data)
}

func (h *UsersHandler) Create(c echo.Context) error {
	name := c.FormValue("name")
	if name == "" {
		return c.Redirect(http.StatusSeeOther, "/users")
	}
	user, err := h.store.CreateUser(c.Request().Context(), name)
	if err != nil {
		users, _ := h.store.ListUsers(c.Request().Context())
		return c.Render(http.StatusOK, "users.html", map[string]any{
			"Users": users,
			"Error": fmt.Sprintf("User %q already exists", name),
		})
	}
	setUserCookie(c, user.ID)
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *UsersHandler) Switch(c echo.Context) error {
	idStr := c.FormValue("user_id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/users")
	}
	setUserCookie(c, userID)
	return c.Redirect(http.StatusSeeOther, "/")
}

func setUserCookie(c echo.Context, userID int64) {
	c.SetCookie(&http.Cookie{
		Name:     appmw.UserIDCookie,
		Value:    strconv.FormatInt(userID, 10),
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
