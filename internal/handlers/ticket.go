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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to insert tickets: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func VectorSearch(c *gin.Context) {
	searchInput := c.Params.ByName("input")
	tickets, err := db.Ticket.VectorSearch(&searchInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to search ticket: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "tickets": tickets})
}
