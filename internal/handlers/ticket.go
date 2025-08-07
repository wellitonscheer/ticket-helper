package handlers

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecervi"
	"github.com/wellitonscheer/ticket-helper/internal/llm"
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

	switch searchType {
	case entireTicketContent:
		tik.SearchTicketBody(c, searchInput)
		return
	case singleMessages:
		tik.SearchSingleMessage(c, searchInput)
		return
	case chunk:
		tik.SearchChunk(c, searchInput)
		return
	case suggestReply:
		tik.SuggestReply(c, searchInput)
		return
	default:
		c.String(http.StatusBadRequest, "invalid search type")
		return
	}
}

func (tik TicketHandlers) SuggestReply(c *gin.Context, searchInput string) {
	ticketChunkService := pgvecervi.NewTicketChunksService(tik.AppCtx)
	ticketEntriesService := pgvecervi.NewTicketEntriesService(tik.AppCtx)

	var context string

	seachedChunks, err := ticketChunkService.SearchComputeScore(types.SearchComputeScoreInput{
		Search:        searchInput,
		Limit:         6,
		RelevantScore: float32(0.8),
	})
	if err != nil {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusInternalServerError,
			LogMsg:  fmt.Sprintf("in SuggestReply failed to SearchComputeScore by searchInput (searchInput=%s): %v", searchInput, err),
			UserMsg: "failed to suggest reply, try again later",
		})
		return
	}

	for _, chunk := range seachedChunks {
		ticketEntries, err := ticketEntriesService.GetByTicketId(chunk.TicketId)
		if err != nil {
			utils.HandleError(c, utils.HandleErrorInput{
				Code:    http.StatusInternalServerError,
				LogMsg:  fmt.Sprintf("in SuggestReply failed to GetByTicketId by chunk ticket id (TicketId=%d): %v", chunk.TicketId, err),
				UserMsg: "failed to suggest reply, try again later",
			})
			return
		}

		if len(ticketEntries) == 0 {
			utils.HandleError(c, utils.HandleErrorInput{
				Code:    http.StatusInternalServerError,
				LogMsg:  fmt.Sprintf("in SuggestReply GetByTicketId returned no values (TicketId=%d)", chunk.TicketId),
				UserMsg: "failed to suggest reply, try again later",
			})
			return
		}

		sort.Slice(ticketEntries, func(i, j int) bool {
			return ticketEntries[i].Ordem > ticketEntries[j].Ordem
		})

		context += fmt.Sprintf("%s:\n", ticketEntries[0].Subject)
		for _, entry := range ticketEntries {
			context += fmt.Sprintf("%s\n", entry.Body)
		}

		context += "\n"
	}

	suggestedReply, err := llm.SuggestReply(&searchInput, &context)
	if err != nil {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusInternalServerError,
			LogMsg:  fmt.Sprintf("in SuggestReply failed to get reply from llm.SuggestReply (searchInput=%s): %v", searchInput, err),
			UserMsg: "failed to suggest reply, try again later",
		})
		return
	}

	c.HTML(http.StatusOK, "suggeted-reply", gin.H{"Reply": suggestedReply})
}

func (tik TicketHandlers) SearchSingleMessage(c *gin.Context, searchInput string) {
	ticketEntriesSer := pgvecervi.NewTicketEntriesService(tik.AppCtx)

	results := []types.TicketVectorSearchResponse{}

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

	c.HTML(http.StatusOK, "results", results)
}

func (tik TicketHandlers) SearchChunk(c *gin.Context, searchInput string) {
	ticketChunkService := pgvecervi.NewTicketChunksService(tik.AppCtx)

	results := []types.TicketVectorSearchResponse{}

	if len(searchInput) > 40 {
		inputChunks := utils.ChunkText(types.ChunkTextInput{
			Text:        searchInput,
			ChunkSize:   40,
			OverlapSize: 10,
		})

		for _, input := range inputChunks {
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
				results = append(results, types.TicketVectorSearchResponse{
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
			results = append(results, types.TicketVectorSearchResponse{
				TicketId: searchedChunk.TicketId,
				Score:    searchedChunk.Distance,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	c.HTML(http.StatusOK, "results", results)
}

func (tik TicketHandlers) SearchTicketBody(c *gin.Context, searchInput string) {
	c.HTML(http.StatusOK, "suggeted-reply", gin.H{"Reply": "nuh uh"})
}
