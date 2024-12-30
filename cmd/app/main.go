package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

	r := gin.Default()

	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/web/static", "./web/static")

	r.GET("/", handlers.Index)
	r.POST("/count", handlers.IndexCount)
	r.POST("/contacts", handlers.IndexCreateContact)
	r.DELETE("/contacts/:id", handlers.IndexDeleteContact)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/user/:name", handlers.UserNew)

	r.GET("/tickets", handlers.TicketInsertAll)
	r.GET("/tickets/search/:input", handlers.TicketVectorSearch)

	r.GET("/tickets/messages/insert-all", handlers.TicketMessagesInsertAll)

	r.GET("/black-tickets/insert-all", handlers.BlackTicketInsertAll)

	r.Run(fmt.Sprintf(":%s", ginPort))
}
