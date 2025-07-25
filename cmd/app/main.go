package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/config"
	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec"
	"github.com/wellitonscheer/ticket-helper/internal/database/sqlite"
	"github.com/wellitonscheer/ticket-helper/internal/handlers"
	"github.com/wellitonscheer/ticket-helper/internal/routes/middleware"
)

func main() {
	conf := config.NewConfig()

	if conf.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	sqliteDb := sqlite.NewSqliteConnection()
	defer sqliteDb.Close()

	pgVec := pgvec.NewPGVectorConnection(conf.PGVector)
	defer pgVec.Close()

	appContext := context.AppContext{
		Config: &conf,
		Sqlite: sqliteDb,
		PGVec:  pgVec,
	}

	sqliteMigrations := sqlite.NewSqliteMigrations(appContext)
	sqliteMigrations.RunMigrations()

	pgvec.RunMigrations(appContext)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/web/static", "./web/static")
	r.StaticFile("/robots.txt", "./robots.txt")

	loginHandlers := handlers.NewLoginHandlers(appContext)
	login := r.Group("/login")
	{
		login.GET("/", loginHandlers.LoginPage)
		login.POST("/send-verification", loginHandlers.SendEmailVefificationCode)
		login.POST("/validate-verification", loginHandlers.LoginWithCode)
	}

	learn := r.Group("/learn")
	{
		learnHandlers := handlers.NewLearnHandlers()

		learn.GET("/", learnHandlers.Learn)
		learn.POST("/count", learnHandlers.Count)
		learn.POST("/contacts", learnHandlers.CreateContact)
		learn.DELETE("/contacts/:id", learnHandlers.DeleteContact)
	}

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(appContext))
	{
		ticketHandlers := handlers.NewTicketHandlers(appContext)

		auth.GET("/", handlers.Index)
		auth.GET("/user/:name", handlers.UserNew)
		auth.POST("/tickets/search", ticketHandlers.TicketVectorSearch)

		auth.GET("/kys", func(c *gin.Context) {
			log.Fatal("Good bye ;-;")
		})
	}

	r.Run(fmt.Sprintf(":%s", conf.Common.GinPort))
}
