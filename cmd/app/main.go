package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/wellitonscheer/ticket-helper/internal/handlers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err.Error())
	}
	ginPort := os.Getenv("GIN_PORT")

	gin.SetMode(gin.DebugMode)
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

	r.GET("/", handlers.Index)

	login := r.Group("/login")
	{
		loginHandlers := handlers.NewLoginHandlers()

		login.GET("/", loginHandlers.LoginPage)
		login.GET("/insert-authorized-emails", loginHandlers.InsertAuthorizedEmails)
		login.POST("/send-verification", loginHandlers.SendEmailVefificationCode)
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
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/user/:name", handlers.UserNew)

	r.GET("/tickets", handlers.TicketInsertAll)
	r.POST("/tickets/search", handlers.TicketVectorSearch)

	r.GET("/tickets/messages/insert-all", handlers.TicketMessagesInsertAll)

	r.GET("/black-tickets/insert-all", handlers.BlackTicketInsertAll)

	r.GET("/kill", func(c *gin.Context) {
		log.Fatal()
	})

	r.Run(fmt.Sprintf(":%s", ginPort))
}
