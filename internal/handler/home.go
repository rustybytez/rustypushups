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

type HomeHandler struct {
	store *store.Store
}

func NewHomeHandler(s *store.Store) *HomeHandler {
	return &HomeHandler{store: s}
}

type homeData struct {
	User          db.User
	Today         string
	TodayShort    string
	TodayCount    int64
	MonthPrefix   string
	MonthTotal    int64
	MonthGoal     int64
	GoalSet       bool
	ProgressPct   int
	DailyNeeded   int64
	DaysRemaining int
	PrevDate      string // empty if 1st of month
	NextDate      string // always set; template hides if it's today
}

func parseClientDate(clientDate string) (today string, t time.Time) {
	if clientDate != "" {
		if parsed, err := time.Parse("2006-01-02", clientDate); err == nil {
			return clientDate, parsed
		}
	}
	now := time.Now()
	return now.Format("2006-01-02"), now
}

func (h *HomeHandler) buildData(c echo.Context, clientDate string) (*homeData, error) {
	user := c.Get(appmw.UserContextKey).(db.User)

	today, t := parseClientDate(clientDate)
	year, month, day := t.Year(), int(t.Month()), t.Day()
	monthPrefix := fmt.Sprintf("%04d-%02d%%", year, month)
	lastDayOfMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	daysRemaining := lastDayOfMonth - day + 1

	// Prev/next navigation
	prevDate := ""
	if day > 1 {
		prevDate = t.AddDate(0, 0, -1).Format("2006-01-02")
	}
	nextDate := t.AddDate(0, 0, 1).Format("2006-01-02")

	ctx := c.Request().Context()

	var todayCount int64
	logEntry, err := h.store.GetDailyLog(ctx, db.GetDailyLogParams{UserID: user.ID, Date: today})
	if err == nil {
		todayCount = logEntry.Count
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	combined, err := h.store.GetCombinedMonthlyTotal(ctx, monthPrefix)
	if err != nil {
		return nil, err
	}
	monthTotal := toInt64(combined)

	var monthGoal int64
	goalSet := false
	goal, err := h.store.GetMonthlyGoal(ctx, db.GetMonthlyGoalParams{
		Year:  int64(year),
		Month: int64(month),
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

	var dailyNeeded int64
	if goalSet && monthGoal > monthTotal && daysRemaining > 0 {
		remaining := monthGoal - monthTotal
		dailyNeeded = (remaining + int64(daysRemaining) - 1) / int64(daysRemaining)
	}

	return &homeData{
		User:          user,
		Today:         today,
		TodayShort:    t.Format("January 2"),
		TodayCount:    todayCount,
		MonthPrefix:   t.Format("January 2006"),
		MonthTotal:    monthTotal,
		MonthGoal:     monthGoal,
		GoalSet:       goalSet,
		ProgressPct:   pct,
		DailyNeeded:   dailyNeeded,
		DaysRemaining: daysRemaining,
		PrevDate:      prevDate,
		NextDate:      nextDate,
	}, nil
}

func (h *HomeHandler) Index(c echo.Context) error {
	clientDate := c.QueryParam("date")
	data, err := h.buildData(c, clientDate)
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

	today, _ := parseClientDate(c.FormValue("date"))

	ctx := c.Request().Context()
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
	data, err := h.buildData(c, today)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "today.html", data)
}

func (h *HomeHandler) Edit(c echo.Context) error {
	user := c.Get(appmw.UserContextKey).(db.User)
	countStr := c.FormValue("count")
	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil || count < 0 {
		count = 0
	}

	today, _ := parseClientDate(c.FormValue("date"))

	err = h.store.UpsertDailyLog(c.Request().Context(), db.UpsertDailyLogParams{
		UserID: user.ID,
		Date:   today,
		Count:  count,
	})
	if err != nil {
		return err
	}
	data, err := h.buildData(c, today)
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
