package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/db"
)

const (
	entireTicketContent = "entire"
	singleMessages      = "message"
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
	searchInput := c.PostForm("search-input")
	searchType := c.PostForm("search-type")

	var tickets db.TicketSearchResults
	if searchType == entireTicketContent {
		ticketService, err := db.NewTicketService()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed create ticket service: %s", err.Error())})
			return
		}

		tickets, err = ticketService.VectorSearch(&searchInput)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to search ticket: %s", err.Error())})
			return
		}
	} else if searchType == singleMessages {
		ticketMessage, err := db.NewTicketMessage()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to get ticket message service: %s", err.Error())})
			return
		}

		tickets, err = ticketMessage.VectorSearch(&searchInput)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to search message ticket: %s", err.Error())})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid search type"})
		return
	}

	c.HTML(http.StatusOK, "results", tickets)
}
