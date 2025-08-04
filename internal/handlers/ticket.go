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

type TicketOccurrence struct {
	Distances []float32
}

type TicketsMostOccurrences map[int]TicketOccurrence

func (tik TicketHandlers) SearchChunk(c *gin.Context, results *[]types.TicketVectorSearchResponse, searchInput string) {
	ticketChunkService := pgvecervi.NewTicketChunksService(tik.AppCtx)

	mostOcc := TicketsMostOccurrences{}

	if len(searchInput) > 40 {
		searchChunks := utils.ChunkText(types.ChunkTextInput{
			Text:        searchInput,
			ChunkSize:   40,
			OverlapSize: 10,
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
				if occ, ok := mostOcc[found.TicketId]; ok {
					mostOcc[found.TicketId] = TicketOccurrence{
						Distances: append(occ.Distances, found.Distance),
					}

				} else {
					mostOcc[found.TicketId] = TicketOccurrence{
						Distances: []float32{found.Distance},
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
			if found.TicketId == 8661 {
				fmt.Printf("\nchunk: %s", searchInput)
				fmt.Printf("bb\n%+v\n", found)
			}
			if occ, ok := mostOcc[found.TicketId]; ok {
				mostOcc[found.TicketId] = TicketOccurrence{
					Distances: append(occ.Distances, found.Distance),
				}

			} else {
				mostOcc[found.TicketId] = TicketOccurrence{
					Distances: []float32{found.Distance},
				}
			}
		}
	}

	var occs []types.TicketVectorSearchResponse
	for ticketId, occ := range mostOcc {
		var biggerDistance float32
		var occMod float32

		for _, distance := range occ.Distances {
			if distance > float32(0.73) {
				occMod = occMod + float32(0.05)
			}
			if distance > biggerDistance {
				biggerDistance = distance
			}
		}

		score := biggerDistance + occMod
		occs = append(occs, types.TicketVectorSearchResponse{TicketId: ticketId, Score: score})
	}

	sort.Slice(occs, func(i, j int) bool {
		return occs[i].Score > occs[j].Score
	})

	*results = append(*results, occs...)
}
