package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/db"
)

func TicketInsertAll(c *gin.Context) {
	err := db.Ticket.InsertAllTickets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to insert tickets: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func TicketVectorSearch(c *gin.Context) {
	searchInput := c.PostForm("search-input")
	searchType := c.PostForm("search-type")
	fmt.Printf("input: %s, type: %s", searchInput, searchType)
	tickets, err := db.Ticket.VectorSearch(&searchInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to search ticket: %s", err.Error())})
		return
	}

	c.HTML(http.StatusOK, "results", tickets)
}
