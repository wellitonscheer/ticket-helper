package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/milvus"
)

func TicketMessagesInsertAll(c *gin.Context) {
	ticketMessage, err := milvus.NewTicketMessage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to get ticket message service: %s", err.Error())})
		return
	}

	err = ticketMessage.InsertAllTickets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to insert ticket messages: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
