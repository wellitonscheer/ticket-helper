package pgvecervi

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"
	"github.com/wellitonscheer/ticket-helper/internal/types"
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
