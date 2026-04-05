-- name: ListUsers :many
SELECT id, name, created_at FROM users ORDER BY name;

-- name: GetUser :one
SELECT id, name, created_at FROM users WHERE id = ?;

-- name: CreateUser :one
INSERT INTO users (name) VALUES (?) RETURNING id, name, created_at;

-- name: UpsertDailyLog :exec
INSERT INTO daily_logs (user_id, date, count)
VALUES (?, ?, ?)
ON CONFLICT(user_id, date) DO UPDATE SET count = excluded.count;

-- name: GetDailyLog :one
SELECT id, user_id, date, count FROM daily_logs
WHERE user_id = ? AND date = ?;

-- name: ListDailyLogs :many
SELECT id, user_id, date, count FROM daily_logs
WHERE user_id = ? AND date LIKE ?
ORDER BY date DESC;

-- name: GetUserMonthlyTotal :one
SELECT COALESCE(SUM(count), 0) AS total FROM daily_logs
WHERE user_id = ? AND date LIKE ?;

-- name: ListAllDailyLogs :many
SELECT id, user_id, date, count FROM daily_logs
WHERE date LIKE ?
ORDER BY date ASC;

-- name: GetCombinedMonthlyTotal :one
SELECT COALESCE(SUM(count), 0) AS total FROM daily_logs
WHERE date LIKE ?;

-- name: UpsertMonthlyGoal :exec
INSERT INTO monthly_goals (year, month, goal)
VALUES (?, ?, ?)
ON CONFLICT(year, month) DO UPDATE SET goal = excluded.goal;

-- name: GetMonthlyGoal :one
SELECT id, year, month, goal FROM monthly_goals
WHERE year = ? AND month = ?;
