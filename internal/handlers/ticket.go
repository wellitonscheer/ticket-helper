package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecervi"
	types "github.com/wellitonscheer/ticket-helper/internal/types/handlers"
	"github.com/wellitonscheer/ticket-helper/internal/utils"
)

const (
	entireTicketContent = "entire"
	singleMessages      = "message"
	suggestReply        = "reply"
)

type TicketHandlers struct {
	AppCtx appContext.AppContext
}

func NewTicketHandlers(appCtx appContext.AppContext) TicketHandlers {
	return TicketHandlers{
		AppCtx: appCtx,
	}
}

func (tik TicketHandlers) TicketVectorSearch(c *gin.Context) {
	searchInput := c.PostForm("search-input")
	searchType := c.PostForm("search-type")

	if searchInput == "" || searchType == "" {
		c.String(http.StatusBadRequest, "invalid search data")
		return
	}

	var results []types.TicketVectorSearchResponse

	if searchType == entireTicketContent {

	} else if searchType == singleMessages {
		ticketEntriesSer := pgvecervi.NewTicketEntriesService(tik.AppCtx)

		tickerEntries, err := ticketEntriesSer.SearchSimilarByText(searchInput)
		if err != nil {
			utils.HandleError(c, utils.HandleErrorInput{
				Code:    http.StatusInternalServerError,
				LogMsg:  fmt.Sprintf("failed to search similar ticket entries by text (searchInput=%s): %v", searchInput, err),
				UserMsg: "failed to search tickets, try again later",
			})
		}

		for _, entry := range tickerEntries {
			results = append(results, types.TicketVectorSearchResponse{TicketId: entry.TicketId, Score: float32(0.42222222)})
		}
	} else if searchType == suggestReply {

	} else {
		c.String(http.StatusBadRequest, "invalid search type")
		return
	}

	c.HTML(http.StatusOK, "results", results)
}
