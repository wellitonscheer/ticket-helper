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
	"github.com/wellitonscheer/ticket-helper/internal/database/sqlite"
	"github.com/wellitonscheer/ticket-helper/internal/handlers"
	"github.com/wellitonscheer/ticket-helper/internal/milvus"
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

	milvus := milvus.NewMilvusConnection(&conf)
	defer milvus.Client.Close()
	defer milvus.Cancel()

	appContext := context.AppContext{
		Config: &conf,
		Sqlite: sqliteDb,
		Milvus: milvus,
	}

	sqliteMigrations := sqlite.NewSqliteMigrations(appContext)
	sqliteMigrations.RunMigrations()

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
		auth.GET("/", handlers.Index)
		auth.GET("/user/:name", handlers.UserNew)
		auth.GET("/tickets", handlers.TicketInsertAll)
		auth.POST("/tickets/search", handlers.TicketVectorSearch)
		auth.GET("/tickets/messages/insert-all", handlers.TicketMessagesInsertAll)
		auth.GET("/black-tickets/insert-all", handlers.BlackTicketInsertAll)

		auth.GET("/kys", func(c *gin.Context) {
			log.Fatal("Good bye ;-;")
		})
	}

	r.Run(fmt.Sprintf(":%s", conf.Common.GinPort))
}
