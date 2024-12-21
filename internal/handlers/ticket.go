package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/db"
)

func Ticket(c *gin.Context) {
	err := db.Ticket.InsertAllTickets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to insert tickets: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, "success")
}
