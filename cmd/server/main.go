package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/ivansantos/rustypushups/internal/handler"
	appmw "github.com/ivansantos/rustypushups/internal/middleware"
	"github.com/ivansantos/rustypushups/internal/store"
	"github.com/ivansantos/rustypushups/web"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "rustypushups.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Fatal("AUTH_TOKEN env var is required")
	}

	s, err := store.New(dsn)
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Renderer = web.NewRenderer()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/manifest.json", func(c echo.Context) error {
		data, err := web.FS.ReadFile("manifest.json")
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "application/manifest+json", data)
	})

	// Login routes — always accessible
	authH := handler.NewAuthHandler(authToken)
	e.GET("/login", authH.LoginPage)
	e.POST("/login", authH.Login)

	// All other routes require the auth token (if AUTH_TOKEN is set)
	app := e.Group("", appmw.TokenRequired(authToken))

	usersH := handler.NewUsersHandler(s)
	app.GET("/users", usersH.List)
	app.POST("/users", usersH.Create)
	app.POST("/users/switch", usersH.Switch)

	protected := app.Group("", appmw.UserRequired(s))

	homeH := handler.NewHomeHandler(s)
	protected.GET("/", homeH.Index)
	protected.POST("/log", homeH.Log)
	protected.POST("/goal", homeH.SetGoal)

	historyH := handler.NewHistoryHandler(s)
	protected.GET("/history", historyH.Index)

	log.Printf("listening on :%s", port)
	log.Fatal(e.Start(":" + port))
}
