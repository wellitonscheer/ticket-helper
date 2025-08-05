package handlers

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecervi"
	"github.com/wellitonscheer/ticket-helper/internal/types"
	"github.com/wellitonscheer/ticket-helper/internal/utils"
)

const (
	singleMessages      = "message"
	chunk               = "chunk"
	entireTicketContent = "entire"
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

	results := []types.TicketVectorSearchResponse{}

	if searchType == entireTicketContent {

	} else if searchType == chunk {
		tik.SearchChunk(c, &results, searchInput)
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
			results = append(results, types.TicketVectorSearchResponse{TicketId: entry.TicketId, Score: entry.Distance})
		}
	} else if searchType == suggestReply {

	} else {
		c.String(http.StatusBadRequest, "invalid search type")
		return
	}

	c.HTML(http.StatusOK, "results", results)
}

func (tik TicketHandlers) SearchChunk(c *gin.Context, results *[]types.TicketVectorSearchResponse, searchInput string) {
	ticketChunkService := pgvecervi.NewTicketChunksService(tik.AppCtx)

	if len(searchInput) > 40 {
		inputChunks := utils.ChunkText(types.ChunkTextInput{
			Text:        searchInput,
			ChunkSize:   40,
			OverlapSize: 10,
		})

		for _, input := range inputChunks {
			fmt.Println(input)
			seachedChunks, err := ticketChunkService.SearchComputeScore(types.SearchComputeScoreInput{
				Search: input,
			})
			if err != nil {
				utils.HandleError(c, utils.HandleErrorInput{
					Code:    http.StatusInternalServerError,
					LogMsg:  fmt.Sprintf("failed to SearchComputeScore by searchInput chunk (searchChunk=%s): %v", input, err),
					UserMsg: "failed to search chunks, try again later",
				})
				return
			}

			for _, searchedChunk := range seachedChunks {
				*results = append(*results, types.TicketVectorSearchResponse{
					TicketId: searchedChunk.TicketId,
					Score:    searchedChunk.Distance,
				})
			}
		}
	} else {
		seachedChunks, err := ticketChunkService.SearchComputeScore(types.SearchComputeScoreInput{
			Search: searchInput,
		})
		if err != nil {
			utils.HandleError(c, utils.HandleErrorInput{
				Code:    http.StatusInternalServerError,
				LogMsg:  fmt.Sprintf("failed to SearchComputeScore chunks by text (searchInput=%s): %v", searchInput, err),
				UserMsg: "failed to search chunks, try again later",
			})
		}

		for _, searchedChunk := range seachedChunks {
			*results = append(*results, types.TicketVectorSearchResponse{
				TicketId: searchedChunk.TicketId,
				Score:    searchedChunk.Distance,
			})
		}
	}

	sort.Slice((*results), func(i, j int) bool {
		return (*results)[i].Score > (*results)[j].Score
	})
}
