package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/db"
)

func BlackTicketInsertAll(c *gin.Context) {
	blackTicket, err := db.NewBlackTicket()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to get black ticket service: %s", err.Error())})
		return
	}

	err = blackTicket.InsertAllTickets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to insert black tickets: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
