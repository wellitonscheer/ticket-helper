package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/milvus"
	"github.com/wellitonscheer/ticket-helper/internal/database/milvus/milservi"
	"github.com/wellitonscheer/ticket-helper/internal/llm"
	"github.com/wellitonscheer/ticket-helper/internal/types"
	"github.com/wellitonscheer/ticket-helper/internal/utils"
)

type TicketHandlers struct {
	milvus     *milvus.MilvusClient
	appContext context.AppContext
}

func NewTicketHandlers(appContext context.AppContext) TicketHandlers {
	return TicketHandlers{
		milvus:     appContext.Milvus,
		appContext: appContext,
	}
}

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

func (t TicketHandlers) TicketVectorSearch(c *gin.Context) {
	searchInput := c.PostForm("search-input")
	searchType := c.PostForm("search-type")

	if searchInput == "" || searchType == "" {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusBadRequest,
			LogMsg:  "invalid search data",
			UserMsg: "invalid search data",
		})
		return
	}

	var tickets types.TicketSearchResults
	if searchType == entireTicketContent {
		ticketService := milservi.NewTicketService(t.appContext)
		
		var err error
		tickets, err = ticketService.VectorSearchTicketsIds(searchInput)
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
