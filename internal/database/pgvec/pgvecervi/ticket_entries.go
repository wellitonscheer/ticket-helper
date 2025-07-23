package pgvecervi

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/wellitonscheer/ticket-helper/internal/client"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"
	"github.com/wellitonscheer/ticket-helper/internal/types"
)

const (
	limitSimilaritySearch int = 10
)

type TicketEntriesService struct {
	Conn   *pgxpool.Pool
	AppCtx appContext.AppContext
}

func NewPGTicketServices(appCtx appContext.AppContext) TicketEntriesService {
	return TicketEntriesService{
		Conn:   appCtx.PGVec,
		AppCtx: appCtx,
	}
}

func (tik TicketEntriesService) Create(ticket pgvecodel.TicketEntry) error {
	ctx := context.Background()
	_, err := tik.Conn.Exec(
		ctx,
		`
			INSERT INTO ticket_entries (type, ticket_id, subject, ordem, poster, body, embedding)
			VALUES ($1, $2, $3, $4, $5, $6, $7);
		`,
		ticket.Type, ticket.TicketId, ticket.Subject, ticket.Ordem, ticket.Poster, ticket.Body, ticket.Embedding,
	)
	if err != nil {
		return fmt.Errorf("failed to create new ticket (ticket=%+v): %v", ticket, err)
	}

	return nil
}

func (tik TicketEntriesService) GetByTicketId(ticketId int) ([]pgvecodel.TicketEntry, error) {
	var entries []pgvecodel.TicketEntry

	err := pgxscan.Select(context.Background(), tik.Conn, &entries, "SELECT * FROM ticket_entries WHERE ticket_id = $1", ticketId)
	if err != nil {
		return entries, fmt.Errorf("failed to get ticket entries by id (ticketId=%d): %v", ticketId, err)
	}

	return entries, nil
}

func (tik TicketEntriesService) GetUniqueByTicketIdAndOrdem(ticketId int, ordem int) (pgvecodel.TicketEntry, error) {
	var entries []pgvecodel.TicketEntry

	err := pgxscan.Select(context.Background(), tik.Conn, &entries, "SELECT * FROM ticket_entries WHERE ticket_id = $1 AND ordem = $2", ticketId, ordem)
	if err != nil {
		return pgvecodel.TicketEntry{}, fmt.Errorf("failed to get unique ticket entry by id and ordem (ticketId=%d): %v", ticketId, err)
	}

	if len(entries) == 0 {
		return pgvecodel.TicketEntry{}, nil
	}

	return entries[0], nil
}

func (tik TicketEntriesService) SearchSimilarByEmbed(embed []float32) ([]pgvecodel.TicketEntry, error) {
	var entries []pgvecodel.TicketEntry

	err := pgxscan.Select(context.Background(), tik.Conn, &entries, "SELECT * FROM ticket_entries ORDER BY embedding <=> $1 LIMIT $2", pgvector.NewVector(embed), limitSimilaritySearch)
	if err != nil {
		return entries, fmt.Errorf("failed ticket entries by embed (embed=%v): %v", embed, err)
	}

	return entries, nil
}

func (tik TicketEntriesService) SearchSimilarByText(text string) ([]pgvecodel.TicketEntry, error) {
	embedInputs := types.Inputs{
		Inputs: []string{text},
	}
	embeddings, err := client.GetTextEmbeddings(tik.AppCtx, &embedInputs)
	if err != nil {
		return []pgvecodel.TicketEntry{}, fmt.Errorf("failed to get text embeddings for the similarity search (text=%s): %v", text, err)
	}

	firstEmbedding := (*embeddings)[0]

	if len(firstEmbedding) == 0 {
		return []pgvecodel.TicketEntry{}, fmt.Errorf("getTextEmbeddings returned no embedding (embeddings=%+v)", embeddings)
	}

	return tik.SearchSimilarByEmbed(firstEmbedding)
}
