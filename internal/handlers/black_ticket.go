package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/milvus"
)

func BlackTicketInsertAll(c *gin.Context) {
	blackTicket, err := milvus.NewBlackTicket()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("failed to get black ticket service: %s", err.Error()))
		return
	}

	err = blackTicket.InsertAllTickets()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("failed to insert black tickets: %s", err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
