package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ivansantos/rustypushups/internal/db"
	appmw "github.com/ivansantos/rustypushups/internal/middleware"
	"github.com/ivansantos/rustypushups/internal/store"
	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
	store *store.Store
}

func NewHomeHandler(s *store.Store) *HomeHandler {
	return &HomeHandler{store: s}
}

type homeData struct {
	User          db.User
	Today         string
	TodayShort    string // "April 5"
	TodayCount    int64
	MonthPrefix   string
	MonthTotal    int64 // combined all users
	MonthGoal     int64
	GoalSet       bool
	ProgressPct   int
	DailyNeeded   int64
	DaysRemaining int
}

func (h *HomeHandler) buildData(c echo.Context) (*homeData, error) {
	user := c.Get(appmw.UserContextKey).(db.User)
	now := time.Now()
	today := now.Format("2006-01-02")
	monthPrefix := now.Format("2006-01") + "%"

	ctx := c.Request().Context()

	var todayCount int64
	logEntry, err := h.store.GetDailyLog(ctx, db.GetDailyLogParams{UserID: user.ID, Date: today})
	if err == nil {
		todayCount = logEntry.Count
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Combined total of all users for the month
	combined, err := h.store.GetCombinedMonthlyTotal(ctx, monthPrefix)
	if err != nil {
		return nil, err
	}
	monthTotal := toInt64(combined)

	// Shared monthly goal
	var monthGoal int64
	goalSet := false
	goal, err := h.store.GetMonthlyGoal(ctx, db.GetMonthlyGoalParams{
		Year:  int64(now.Year()),
		Month: int64(now.Month()),
	})
	if err == nil {
		monthGoal = goal.Goal
		goalSet = true
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	pct := 0
	if goalSet && monthGoal > 0 {
		pct = int(monthTotal * 100 / monthGoal)
		if pct > 100 {
			pct = 100
		}
	}

	lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location())
	daysRemaining := lastDay.Day() - now.Day() + 1

	var dailyNeeded int64
	if goalSet && monthGoal > monthTotal && daysRemaining > 0 {
		remaining := monthGoal - monthTotal
		dailyNeeded = (remaining + int64(daysRemaining) - 1) / int64(daysRemaining)
	}

	return &homeData{
		User:          user,
		Today:         today,
		TodayShort:    now.Format("January 2"),
		TodayCount:    todayCount,
		MonthPrefix:   now.Format("January 2006"),
		MonthTotal:    monthTotal,
		MonthGoal:     monthGoal,
		GoalSet:       goalSet,
		ProgressPct:   pct,
		DailyNeeded:   dailyNeeded,
		DaysRemaining: daysRemaining,
	}, nil
}

func (h *HomeHandler) Index(c echo.Context) error {
	data, err := h.buildData(c)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "home.html", data)
}

func (h *HomeHandler) Log(c echo.Context) error {
	user := c.Get(appmw.UserContextKey).(db.User)
	countStr := c.FormValue("count")
	add, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil || add < 0 {
		add = 0
	}

	ctx := c.Request().Context()
	today := time.Now().Format("2006-01-02")

	var current int64
	existing, err := h.store.GetDailyLog(ctx, db.GetDailyLogParams{UserID: user.ID, Date: today})
	if err == nil {
		current = existing.Count
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	err = h.store.UpsertDailyLog(ctx, db.UpsertDailyLogParams{
		UserID: user.ID,
		Date:   today,
		Count:  current + add,
	})
	if err != nil {
		return err
	}
	data, err := h.buildData(c)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "today.html", data)
}

func (h *HomeHandler) SetGoal(c echo.Context) error {
	goalStr := c.FormValue("goal")
	goal, err := strconv.ParseInt(goalStr, 10, 64)
	if err != nil || goal < 0 {
		goal = 0
	}
	now := time.Now()
	err = h.store.UpsertMonthlyGoal(c.Request().Context(), db.UpsertMonthlyGoalParams{
		Year:  int64(now.Year()),
		Month: int64(now.Month()),
		Goal:  goal,
	})
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/users")
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case float64:
		return int64(x)
	case []byte:
		n, _ := strconv.ParseInt(string(x), 10, 64)
		return n
	}
	return 0
}
