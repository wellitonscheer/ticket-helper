package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

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

type TicketOccurrence struct {
	TicketId int
	Score    float32
	Distance float32
}

type TicketsMostOccurrences map[int]TicketOccurrence

func (tik TicketHandlers) SearchChunk(c *gin.Context, results *[]types.TicketVectorSearchResponse, searchInput string) {
	ticketChunkService := pgvecervi.NewTicketChunksService(tik.AppCtx)

	mostOcc := TicketsMostOccurrences{}

	if len(searchInput) > 20 {
		searchChunks := utils.ChunkText(types.ChunkTextInput{
			Text:        searchInput,
			ChunkSize:   20,
			OverlapSize: 7,
		})

		for _, searchChunk := range searchChunks {
			seachedChunks, err := ticketChunkService.SearchSimilarByText(searchChunk)
			if err != nil {
				utils.HandleError(c, utils.HandleErrorInput{
					Code:    http.StatusInternalServerError,
					LogMsg:  fmt.Sprintf("failed to search similar searchInput chunk (searchChunk=%s): %v", searchChunk, err),
					UserMsg: "failed to search chunks, try again later",
				})
				return
			}

			for _, found := range seachedChunks {
				if mostOcc[found.TicketId] == (TicketOccurrence{}) {
					mostOcc[found.TicketId] = TicketOccurrence{
						TicketId: found.TicketId,
						Score:    ((1 + found.Distance) * 2),
						Distance: found.Distance,
					}
				} else {
					mostOcc[found.TicketId] = TicketOccurrence{
						TicketId: found.TicketId,
						Score:    mostOcc[found.TicketId].Score + ((1 + found.Distance) * 2),
						Distance: mostOcc[found.TicketId].Distance,
					}
				}
			}
		}

		searchInputWords := strings.Split(searchInput, " ")
		for _, searchWord := range searchInputWords {
			if len(searchWord) < 5 {
				continue
			}

			seachedWord, err := ticketChunkService.SearchSimilarByText(searchWord)
			if err != nil {
				utils.HandleError(c, utils.HandleErrorInput{
					Code:    http.StatusInternalServerError,
					LogMsg:  fmt.Sprintf("failed to search similar searchInput word (searchWord=%s): %v", searchWord, err),
					UserMsg: "failed to search chunks, try again later",
				})
				return
			}

			for _, found := range seachedWord {
				if mostOcc[found.TicketId] == (TicketOccurrence{}) {
					mostOcc[found.TicketId] = TicketOccurrence{
						TicketId: found.TicketId,
						Score:    ((1 + found.Distance) / 2),
						Distance: found.Distance,
					}
				} else {
					mostOcc[found.TicketId] = TicketOccurrence{
						TicketId: found.TicketId,
						Score:    mostOcc[found.TicketId].Score + ((1 + found.Distance) / 2),
						Distance: mostOcc[found.TicketId].Distance,
					}
				}
			}
		}
	} else {
		ticketChunks, err := ticketChunkService.SearchSimilarByText(searchInput)
		if err != nil {
			utils.HandleError(c, utils.HandleErrorInput{
				Code:    http.StatusInternalServerError,
				LogMsg:  fmt.Sprintf("failed to search similar ticket chunks by text (searchInput=%s): %v", searchInput, err),
				UserMsg: "failed to search chunks, try again later",
			})
		}

		for _, found := range ticketChunks {
			if mostOcc[found.TicketId] == (TicketOccurrence{}) {
				mostOcc[found.TicketId] = TicketOccurrence{
					TicketId: found.TicketId,
					Score:    1,
					Distance: found.Distance,
				}
			} else {
				mostOcc[found.TicketId] = TicketOccurrence{
					TicketId: found.TicketId,
					Score:    mostOcc[found.TicketId].Score + 1,
					Distance: mostOcc[found.TicketId].Distance,
				}
			}
		}
	}

	var occs []TicketOccurrence
	for _, occ := range mostOcc {
		occs = append(occs, occ)
	}

	sort.Slice(occs, func(i, j int) bool {
		return occs[i].Score > occs[j].Score
	})

	for _, chunk := range occs {
		*results = append(*results, types.TicketVectorSearchResponse{TicketId: chunk.TicketId, Score: chunk.Distance})
	}
}
