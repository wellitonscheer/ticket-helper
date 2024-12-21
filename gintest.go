package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/wellitonscheer/ticket-helper/db"
)

var dbMock = make(map[string]string)

func FunGin() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err.Error())
	}
	ginPort := os.Getenv("GIN_PORT")

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		err := db.User.Save(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create user: %s", err.Error())})
			return
		}
		value, ok := dbMock[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	r.GET("/tickets", func(c *gin.Context) {
		err := db.Ticket.InsertAllTickets()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to insert tickets: %s", err.Error())})
			return
		}

		c.JSON(http.StatusOK, "success")
	})

	r.Run(fmt.Sprintf(":%s", ginPort))
}
