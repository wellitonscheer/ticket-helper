package pgvecervi

import (
	"context"
	"fmt"
	"sort"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/wellitonscheer/ticket-helper/internal/client"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"
	"github.com/wellitonscheer/ticket-helper/internal/types"
)

const (
	chunkLimitSimilaritySearch = 50
)

type TicketChunksService struct {
	Conn   *pgxpool.Pool
	AppCtx appContext.AppContext
}

func NewTicketChunksService(appCtx appContext.AppContext) TicketChunksService {
	return TicketChunksService{
		Conn:   appCtx.PGVec,
		AppCtx: appCtx,
	}
}

func (chu TicketChunksService) Create(chunk pgvecodel.TicketChunk) error {
	sqlStm := `
		INSERT INTO ticket_chunks (type, ticket_id, subject, ordem, poster, chunk, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err := chu.Conn.Exec(
		context.Background(),
		sqlStm,
		chunk.Type, chunk.TicketId, chunk.Subject, chunk.Ordem, chunk.Poster, chunk.Chunk, chunk.Embedding,
	)
	if err != nil {
		return fmt.Errorf("failed to create new ticket chunk (chunk=%+v): %v", chunk, err)
	}

	return nil
}

func (chu TicketChunksService) Get(filters types.TicketChunkGetInputFilters) ([]pgvecodel.TicketChunk, error) {
	var ticketChunks []pgvecodel.TicketChunk
	if !filters.IsValid() {
		return ticketChunks, fmt.Errorf("invalid get ticket chunks filters input (filters=%+v)", filters)
	}

	sqlStm := "SELECT * FROM ticket_chunks WHERE"
	for i, col := range filters.Columns {
		if i == 0 {
			sqlStm = sqlStm + fmt.Sprintf(" %s = $%d", col, i+1)
		} else {
			sqlStm = sqlStm + fmt.Sprintf(" AND %s = $%d", col, i+1)
		}
	}

	err := pgxscan.Select(context.Background(), chu.Conn, &ticketChunks, sqlStm, filters.Values...)
	if err != nil {
		return ticketChunks, fmt.Errorf("failed to get ticket chunks (sqlStm=%s): %v", sqlStm, err)
	}

	if len(ticketChunks) == 0 {
		return ticketChunks, nil
	}

	return ticketChunks, nil
}

func (chu TicketChunksService) SearchSimilarByEmbed(embed []float32) ([]pgvecodel.TicketChunkSimilaritySearch, error) {
	var chunks []pgvecodel.TicketChunkSimilaritySearch

	sqlStm := "SELECT id, type, ticket_id, subject, ordem, poster, chunk, 1 - (embedding <=> $1) AS distance FROM ticket_chunks ORDER BY distance DESC LIMIT $2"
	err := pgxscan.Select(context.Background(), chu.Conn, &chunks, sqlStm, pgvector.NewVector(embed), chunkLimitSimilaritySearch)
	if err != nil {
		return chunks, fmt.Errorf("failed to search ticket chunks by embed (embed=%v): %v", embed, err)
	}

	return chunks, nil
}

func (chu TicketChunksService) SearchSimilarByText(text string) ([]pgvecodel.TicketChunkSimilaritySearch, error) {
	embedding, err := client.GetSingleTextEmbedding(chu.AppCtx, text)
	if err != nil {
		return []pgvecodel.TicketChunkSimilaritySearch{}, fmt.Errorf("failed to get chunk text embeddings for the similarity search (text=%s): %v", text, err)
	}

	return chu.SearchSimilarByEmbed(embedding)
}

type TicketOccurrence struct {
	Distances     []float32
	TopDistance   float32
	OccMod        float32
	ComputedScore float32
}

type TicketsMostOccurrences map[int]TicketOccurrence

// score = max(similarity) + 0.05 * count(similarities > 0.73)
func (chu TicketChunksService) SearchComputeScore(input types.SearchComputeScoreInput) ([]pgvecodel.TicketChunkSimilaritySearch, error) {
	if input == (types.SearchComputeScoreInput{}) {
		return []pgvecodel.TicketChunkSimilaritySearch{}, fmt.Errorf("empty SearchComputeScore input")
	}

	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.RelevantScore == 0 {
		input.RelevantScore = float32(0.73)
	}

	mostOcc := TicketsMostOccurrences{}
	mapTickets := make(map[int]pgvecodel.TicketChunkSimilaritySearch)

	seachedChunks, err := chu.SearchSimilarByText(input.Search)
	if err != nil {
		return []pgvecodel.TicketChunkSimilaritySearch{}, fmt.Errorf("failed SearchSimilarByText in SearchComputeScore (input=%s): %v", input.Search, err)
	}

	for _, found := range seachedChunks {
		mapTickets[found.TicketId] = found

		lastOcc := mostOcc[found.TicketId]

		currentOcc := TicketOccurrence{
			TopDistance: lastOcc.TopDistance,
			OccMod:      lastOcc.OccMod,
		}

		if found.Distance > lastOcc.TopDistance {
			currentOcc.TopDistance = found.Distance
		}
		if found.Distance > input.RelevantScore {
			currentOcc.OccMod = lastOcc.OccMod + float32(0.05)
		}

		currentOcc.Distances = append(lastOcc.Distances, found.Distance)

		currentOcc.ComputedScore = currentOcc.TopDistance + currentOcc.OccMod

		mostOcc[found.TicketId] = currentOcc
	}

	calculadedChunks := []pgvecodel.TicketChunkSimilaritySearch{}
	for ticketId, occ := range mostOcc {
		calculated := mapTickets[ticketId]
		calculated.Distance = occ.ComputedScore
		calculadedChunks = append(calculadedChunks, calculated)
	}

	sort.Slice(calculadedChunks, func(i, j int) bool {
		return calculadedChunks[i].Distance > calculadedChunks[j].Distance
	})

	if len(calculadedChunks) > input.Limit {
		calculadedChunks = calculadedChunks[:input.Limit]
	}

	return calculadedChunks, nil
}
