package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/db"
)

func TicketInsertAll(c *gin.Context) {
	ticketService, err := db.NewTicketService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed create ticket service: %s", err.Error())})
		return
	}

	err = ticketService.InsertAllTickets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to insert tickets: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func TicketVectorSearch(c *gin.Context) {
	ticketService, err := db.NewTicketService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed create ticket service: %s", err.Error())})
		return
	}

	searchInput := c.PostForm("search-input")
	_ = c.PostForm("search-type")

	tickets, err := ticketService.VectorSearch(&searchInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to search ticket: %s", err.Error())})
		return
	}

	c.HTML(http.StatusOK, "results", tickets)
}
