package handler

import (
	"net/http"
	"time"

	"github.com/ivansantos/rustypushups/internal/db"
	appmw "github.com/ivansantos/rustypushups/internal/middleware"
	"github.com/ivansantos/rustypushups/internal/store"
	"github.com/labstack/echo/v4"
)

type HistoryHandler struct {
	store *store.Store
}

func NewHistoryHandler(s *store.Store) *HistoryHandler {
	return &HistoryHandler{store: s}
}

type historyRow struct {
	Date   string
	Counts []int64 // one per user, same order as Users
}

func (h *HistoryHandler) Index(c echo.Context) error {
	ctx := c.Request().Context()
	now := time.Now()
	monthPrefix := now.Format("2006-01") + "%"

	users, err := h.store.ListUsers(ctx)
	if err != nil {
		return err
	}

	logs, err := h.store.ListAllDailyLogs(ctx, monthPrefix)
	if err != nil {
		return err
	}

	// Build map[date][userID]count
	logMap := make(map[string]map[int64]int64)
	for _, l := range logs {
		if logMap[l.Date] == nil {
			logMap[l.Date] = make(map[int64]int64)
		}
		logMap[l.Date][l.UserID] = l.Count
	}

	// Generate all days of the month
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location())

	rows := make([]historyRow, 0, lastDay.Day())
	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		counts := make([]int64, len(users))
		for i, u := range users {
			counts[i] = logMap[date][u.ID]
		}
		rows = append(rows, historyRow{Date: date, Counts: counts})
	}

	user := c.Get(appmw.UserContextKey).(db.User)
	return c.Render(http.StatusOK, "history.html", map[string]any{
		"User":  user,
		"Month": now.Format("January 2006"),
		"Users": users,
		"Rows":  rows,
	})
}
