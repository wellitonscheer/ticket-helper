package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/llm"
	"github.com/wellitonscheer/ticket-helper/internal/milvus"
)

const (
	entireTicketContent = "entire"
	singleMessages      = "message"
	suggestReply        = "reply"
)

func TicketInsertAll(c *gin.Context) {
	ticketService, err := milvus.NewTicketService()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("failed create ticket service: %s", err.Error()))
		return
	}

	err = ticketService.InsertAllTickets()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("failed to insert tickets: %s", err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

func TicketVectorSearch(c *gin.Context) {
	searchInput := c.PostForm("search-input")
	searchType := c.PostForm("search-type")

	if searchInput == "" || searchType == "" {
		c.String(http.StatusBadRequest, "invalid search data")
		return
	}

	var tickets milvus.TicketSearchTicketsIdsResults
	if searchType == entireTicketContent {
		ticketService, err := milvus.NewTicketService()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("failed create ticket service: %s", err.Error()))
			return
		}

		tickets, err = ticketService.VectorSearchTicketsIds(&searchInput)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("failed to search ticket: %s", err.Error()))
			return
		}
	} else if searchType == singleMessages {
		ticketMessage, err := milvus.NewTicketMessage()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("failed to get ticket message service: %s", err.Error()))
			return
		}

		tickets, err = ticketMessage.VectorSearch(&searchInput)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("failed to search message ticket: %s", err.Error()))
			return
		}
	} else if searchType == suggestReply {
		suggested, err := llm.SuggestReply(&searchInput)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("failed to suggest reply: %s", err.Error()))
			return
		}

		c.HTML(http.StatusOK, "suggeted-reply", gin.H{"Reply": suggested})
		return
	} else {
		c.String(http.StatusBadRequest, "invalid search type")
		return
	}

	c.HTML(http.StatusOK, "results", tickets)
}
